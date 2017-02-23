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
	"fmt"

	// TODO: [ben] swap out go-events with eris-db/event (currently unused)
	events "github.com/tendermint/go-events"

	"github.com/eris-ltd/eris-db/config"
	"github.com/eris-ltd/eris-db/consensus"
	"github.com/eris-ltd/eris-db/definitions"
	"github.com/eris-ltd/eris-db/event"
	"github.com/eris-ltd/eris-db/manager"
	// rpc_v0 is carried over from Eris-DBv0.11 and before on port 1337
	rpc_v0 "github.com/eris-ltd/eris-db/rpc/v0"
	// rpc_tendermint is carried over from Eris-DBv0.11 and before on port 46657

	"github.com/eris-ltd/eris-db/logging"
	"github.com/eris-ltd/eris-db/logging/loggers"
	rpc_tendermint "github.com/eris-ltd/eris-db/rpc/tendermint/core"
	"github.com/eris-ltd/eris-db/server"
)

// Core is the high-level structure
type Core struct {
	chainId        string
	evsw           *events.EventSwitch
	pipe           definitions.Pipe
	tendermintPipe definitions.TendermintPipe
}

func NewCore(chainId string,
	consensusConfig *config.ModuleConfig,
	managerConfig *config.ModuleConfig,
	logger loggers.InfoTraceLogger) (*Core, error) {
	// start new event switch, TODO: [ben] replace with eris-db/event
	evsw := events.NewEventSwitch()
	evsw.Start()
	logger = logging.WithScope(logger, "Core")

	// start a new application pipe that will load an application manager
	pipe, err := manager.NewApplicationPipe(managerConfig, evsw, logger,
		consensusConfig.Version)
	if err != nil {
		return nil, fmt.Errorf("Failed to load application pipe: %v", err)
	}
	logging.TraceMsg(logger, "Loaded pipe with application manager")
	// pass the consensus engine into the pipe
	if e := consensus.LoadConsensusEngineInPipe(consensusConfig, pipe); e != nil {
		return nil, fmt.Errorf("Failed to load consensus engine in pipe: %v", e)
	}
	tendermintPipe, err := pipe.GetTendermintPipe()
	if err != nil {
		logging.TraceMsg(logger, "Tendermint gateway not supported by manager",
			"manager-version", managerConfig.Version)
	}
	return &Core{
		chainId:        chainId,
		evsw:           evsw,
		pipe:           pipe,
		tendermintPipe: tendermintPipe,
	}, nil
}

//------------------------------------------------------------------------------
// Explicit switch that can later be abstracted into an `Engine` definition
// where the Engine defines the explicit interaction of a specific application
// manager with a consensus engine.
// TODO: [ben] before such Engine abstraction,
// think about many-manager-to-one-consensus

//------------------------------------------------------------------------------
// Server functions
// NOTE: [ben] in phase 0 we exactly take over the full server architecture
// from Eris-DB and Tendermint; This is a draft and will be overhauled.

func (core *Core) NewGatewayV0(config *server.ServerConfig) (*server.ServeProcess,
	error) {
	codec := &rpc_v0.TCodec{}
	eventSubscriptions := event.NewEventSubscriptions(core.pipe.Events())
	// The services.
	tmwss := rpc_v0.NewErisDbWsService(codec, core.pipe)
	tmjs := rpc_v0.NewErisDbJsonService(codec, core.pipe, eventSubscriptions)
	// The servers.
	jsonServer := rpc_v0.NewJsonRpcServer(tmjs)
	restServer := rpc_v0.NewRestServer(codec, core.pipe, eventSubscriptions)
	wsServer := server.NewWebSocketServer(config.WebSocket.MaxWebSocketSessions,
		tmwss)
	// Create a server process.
	proc, err := server.NewServeProcess(config, jsonServer, restServer, wsServer)
	if err != nil {
		return nil, fmt.Errorf("Failed to load gateway: %v", err)
	}

	return proc, nil
}

func (core *Core) NewGatewayTendermint(config *server.ServerConfig) (
	*rpc_tendermint.TendermintWebsocketServer, error) {
	if core.tendermintPipe == nil {
		return nil, fmt.Errorf("No Tendermint pipe has been initialised for Tendermint gateway.")
	}
	return rpc_tendermint.NewTendermintWebsocketServer(config,
		core.tendermintPipe, core.evsw)
}
