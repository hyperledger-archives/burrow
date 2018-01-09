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
	"sync"

	"fmt"

	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/consensus/tendermint/query"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
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

// Kernel is the root structure of Burrow
type Kernel struct {
	eventSwitch     events.EventSwitch
	tmNode          *node.Node
	service         rpc.Service
	serverLaunchers []ServerLauncher
	listeners       []net.Listener
	logger          logging_types.InfoTraceLogger
	shutdownNotify  chan struct{}
	shutdownOnce    sync.Once
}

type ServerLauncher struct {
	Name   string
	Launch func(rpc.Service) (net.Listener, error)
}

func NewKernel(privValidator tm_types.PrivValidator, genesisDoc *genesis.GenesisDoc, tmConf *tm_config.Config,
	rpcConfig *rpc.RPCConfig, logger logging_types.InfoTraceLogger) (*Kernel, error) {

	events.NewEventSwitch().Start()
	logger = logging.WithScope(logger, "NewKernel")

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
	// view of sequence values on the node that a client is communicating with.
	// Since we don't currently execute EVM code in the checker possible conflicts are limited to account creation
	// which increments the creator's account Sequence and SendTxs
	service := rpc.NewService(state, eventEmitter, nameReg, blockchain, transactor, query.NewNodeView(tmNode, txCodec),
		logger)

	servers := []ServerLauncher{
		{
			Name: "TM",
			Launch: func(service rpc.Service) (net.Listener, error) {
				return tm.StartServer(service, "/websocket", rpcConfig.TM.ListenAddress, eventEmitter, logger)
			},
		},
	}

	return &Kernel{
		eventSwitch:     eventEmitter,
		tmNode:          tmNode,
		service:         service,
		serverLaunchers: servers,
		logger:          logger,
		shutdownNotify:  make(chan struct{}),
	}, nil
}

// Boot the kernel starting Tendermint and RPC layers
func (kern *Kernel) Boot() error {
	_, err := kern.tmNode.Start()
	if err != nil {
		return fmt.Errorf("error starting Tendermint node: %v", err)
	}
	for _, launcher := range kern.serverLaunchers {
		listener, err := launcher.Launch(kern.service)
		if err != nil {
			return fmt.Errorf("error launching %s server", launcher.Name)
		}

		kern.listeners = append(kern.listeners, listener)
	}
	go kern.supervise()
	return nil
}

// Wait for a graceful shutdown
func (kern *Kernel) WaitForShutdown() {
	// Supports multiple goroutines waiting for shutdown since channel is closed
	<-kern.shutdownNotify
}

// Supervise kernel once booted
func (kern *Kernel) supervise() {
	// TODO: Consider capturing kernel panics from boot and sending them here via a channel where we could
	// perform disaster restarts of the kernel; rejoining the network as if we were a new node.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals
	kern.Shutdown()
}

// Stop the kernel allowing for a graceful shutdown of components in order
func (kern *Kernel) Shutdown() (err error) {
	kern.shutdownOnce.Do(func() {
		logger := logging.WithScope(kern.logger, "Shutdown")
		logging.InfoMsg(logger, "Attempting graceful shutdown...")
		logging.InfoMsg(logger, "Shutting down listeners")
		for _, listener := range kern.listeners {
			err = listener.Close()
		}
		logging.InfoMsg(logger, "Shutting down Tendermint node")
		kern.tmNode.Stop()
		logging.InfoMsg(logger, "Shutdown complete")
		close(kern.shutdownNotify)
	})
	return
}
