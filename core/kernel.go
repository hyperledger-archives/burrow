// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"net"
	"os"
	"os/signal"

	"time"

	acm "github.com/hyperledger/burrow/account"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/consensus/tendermint/query"
	"github.com/hyperledger/burrow/consensus/tendermint/validator"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/lifecycle"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/tm"
	"github.com/hyperledger/burrow/txs"
	tm_config "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	tm_types "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/events"
)

const GenesisAccountBalance = uint64(1 << 20)

// Kernel is the root structure of Burrow
type Kernel struct {
	eventSwitch events.EventSwitch
	tmNode      *node.Node
	service     rpc.Service
	listeners   []net.Listener
	logger      logging_types.InfoTraceLogger
	shutdownCh  chan struct{}
}

func NewKernel(privValidator tm_types.PrivValidator, genesisDoc *genesis.GenesisDoc, tmConf *tm_config.Config,
	logger logging_types.InfoTraceLogger) (*Kernel, error) {

	events.NewEventSwitch().Start()
	logger = logging.WithScope(logger, "Kernel")

	stateDB := dbm.NewDB("burrow_state", dbm.GoLevelDBBackendStr, tmConf.DBDir())
	state := execution.MakeGenesisState(stateDB, genesisDoc)
	state.Save()

	blockchain := bcm.NewBlockchain(genesisDoc)
	evmEvents := event.NewEmitter(logger)

	tmGenesisDoc := tendermint.DeriveGenesisDoc(genesisDoc)
	checker := execution.NewBatchChecker(state, tmGenesisDoc.ChainID, blockchain, logger)
	committer := execution.NewBatchCommitter(state, tmGenesisDoc.ChainID, blockchain, evmEvents, logger)
	tmNode, err := tendermint.NewNode(tmConf, privValidator, tmGenesisDoc, blockchain, checker, committer, logger)
	if err != nil {
		return nil, err
	}
	// Multiplex Tendermint and EVM events
	eventEmitter := event.Multiplex(evmEvents, event.WrapEventSwitch(tmNode.EventSwitch(), logger))

	txCodec := txs.NewGoWireCodec()
	nameReg := execution.NewNameReg(state, blockchain)

	transactor := execution.NewTransactor(blockchain, state, eventEmitter,
		tendermint.BroadcastTxAsyncFunc(tmNode, txCodec), logger)

	// TODO: consider whether we need to be more explicit about pre-commit (check cache) versus committed (state) values
	// Note we pass the checker as the StateIterable to NewService which means the RPC layers will query the check
	// cache state. This is in line with previous behaviour of Burrow and chiefly serves to get provide a pre-commit
	// view of nonce values on the node that a client is communicating with.
	// Since we don't currently execute EVM code in the checker possible conflicts are limited to account creation
	// which increments the creator's account Sequence and SendTxs
	service := rpc.NewService(state, eventEmitter, nameReg, blockchain, transactor, query.NewNodeView(tmNode, txCodec),
		logger)

	return &Kernel{
		eventSwitch: eventEmitter,
		tmNode:      tmNode,
		service:     service,
		logger:      logger,
		shutdownCh:  make(chan struct{}),
	}, nil
}

func NewGenesisKernel() (*Kernel, error) {
	genesisPrivateAccount := acm.GeneratePrivateAccount()
	genesisAccount := acm.FromAddressable(genesisPrivateAccount)
	genesisAccount.AddToBalance(GenesisAccountBalance)
	genesisValidator := acm.AsValidator(genesisAccount)
	privValidator := validator.NewPrivValidatorMemory(genesisPrivateAccount, genesisPrivateAccount)
	genesisDoc := genesis.MakeGenesisDocFromAccounts("GenesisChain", nil, time.Now(),
		map[string]acm.Account{
			"genesisAccount": genesisAccount,
		},
		map[string]acm.Validator{
			"genesisValidator": genesisValidator,
		})
	tmConf := tm_config.DefaultConfig()
	logger, _ := lifecycle.NewStdErrLogger()
	return NewKernel(privValidator, genesisDoc, tmConf, logger)
}

func (kern *Kernel) Boot() error {
	kern.tmNode.Start()
	tmListener, err := tm.StartServer(kern.service, "/websocket", ":46657", kern.eventSwitch,
		tendermint.NewLogger(kern.logger))
	if err != nil {
		return err
	}
	kern.listeners = append(kern.listeners, tmListener)

	go kern.supervise()
	return nil
}

func (kern *Kernel) WaitForShutdown() {
	<-kern.shutdownCh
}

func (kern *Kernel) supervise() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals
	kern.Shutdown()
}

// Stop the core allowing for a graceful shutdown of component in order.
func (kern *Kernel) Shutdown() {
	logger := logging.WithScope(kern.logger, "Shutdown")
	logging.InfoMsg(logger, "Attempting graceful shutdown...")
	logging.InfoMsg(logger, "Shutting down listeners")
	for _, listener := range kern.listeners {
		listener.Close()
	}
	logging.InfoMsg(logger, "Shutting down Tendermint node")
	kern.tmNode.Stop()
	logging.InfoMsg(logger, "Shutdown complete")
	close(kern.shutdownCh)
}
