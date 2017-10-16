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

	acm "github.com/hyperledger/burrow/account"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/consensus/tendermint/query"
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
}

func NewKernel(privValidator tm_types.PrivValidator, genesisDoc *genesis.GenesisDoc, tmConf *tm_config.Config,
	logger logging_types.InfoTraceLogger) (*Kernel, error) {

	events.NewEventSwitch().Start()
	logger = logging.WithScope(logger, "Kernel")

	stateDB := dbm.NewDB("burrow_state", dbm.GoLevelDBBackendStr, tmConf.DBDir())
	state := execution.MakeGenesisState(stateDB, genesisDoc)
	state.Save()

	blockchain := bcm.NewBlockchain(genesisDoc)
	evmEvents := event.NewEvents(events.NewEventSwitch(), logger)

	tmGenesisDoc := tendermint.DeriveGenesisDoc(genesisDoc)
	checker := execution.NewBatchChecker(state, tmGenesisDoc.ChainID, blockchain, logger)
	committer := execution.NewBatchCommitter(state, tmGenesisDoc.ChainID, blockchain, evmEvents, logger)
	tmNode, err := tendermint.NewNode(tmConf, privValidator, tmGenesisDoc, blockchain, checker, committer, logger)
	if err != nil {
		return nil, err
	}

	// Multiplex Tendermint and EVM events
	eventEmitter := event.Multiplex(evmEvents, event.NewEvents(tmNode.EventSwitch(), logger))

	txCodec := txs.NewGoWireCodec()
	nameReg := execution.NewNameReg(state, blockchain)

	transactor := execution.NewTransactor(blockchain, state, eventEmitter,
		tendermint.BroadcastTxAsyncFunc(tmNode, txCodec), logger)

	service := rpc.NewService(state, eventEmitter, nameReg, blockchain, transactor, query.NewNodeView(tmNode, txCodec),
		logger)

	return &Kernel{
		eventSwitch: eventEmitter,
		tmNode:      tmNode,
		service:     service,
		logger:      logger,
	}, nil
}

func NewGenesisKernel() (*Kernel, error) {
	genesisPrivateAccount := acm.GeneratePrivateAccount()
	genesisAccount := acm.FromAddressable(genesisPrivateAccount)
	genesisAccount.AddToBalance(GenesisAccountBalance)
	genesisValidator := acm.AsValidator(genesisAccount)
	privValidator := tendermint.NewPrivValidatorMemory(genesisPrivateAccount)
	genesisDoc := genesis.MakeGenesisDocFromAccounts("GenesisChain", nil,
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

func (kern *Kernel) supervise() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals
	kern.Shutdown()
}

// Stop the core allowing for a graceful shutdown of component in order.
func (kern *Kernel) Shutdown() {
	for _, listener := range kern.listeners {
		listener.Close()
	}
	kern.tmNode.Stop()
}
