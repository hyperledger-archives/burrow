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

	// TODO: [ben] swap out go-events with burrow/event (currently unused)
	events "github.com/tendermint/tmlibs/events"

	"github.com/hyperledger/burrow/event"
	// rpc_v0 is carried over from burrowv0.11 and before on port 1337
	rpc_v0 "github.com/hyperledger/burrow/rpc/v0"
	// rpc_tendermint is carried over from burrowv0.11 and before on port 46657

	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	rpc_tendermint "github.com/hyperledger/burrow/rpc/tm/core"
	"github.com/hyperledger/burrow/server"
	"github.com/tendermint/tendermint/node"
)

// Core is the top-level structure of Burrow
type Core struct {
	chainId        string
	eventSwitch    events.EventSwitch
	tendermintNode *node.Node
	logger         logging_types.InfoTraceLogger
}

func NewCore(chainId string, logger logging_types.InfoTraceLogger) (*Core, error) {
	// start new event switch, TODO: [ben] replace with burrow/event
	evsw := events.NewEventSwitch()
	evsw.Start()
	logger = logging.WithScope(logger, "Core")

	return &Core{
		chainId:     chainId,
		eventSwitch: evsw,
		logger:      logger,
	}, nil
}

func (core *Core) NewGatewayV0(config *server.ServerConfig) (*server.ServeProcess,
	error) {
	codec := &rpc_v0.TCodec{}
	eventSubscriptions := event.NewEventSubscriptions(core.pipe.Events())
	// The services.
	tmwss := rpc_v0.NewBurrowWsService(codec, core.pipe)
	tmjs := rpc_v0.NewBurrowJsonService(codec, core.pipe, eventSubscriptions)
	// The servers.
	jsonServer := rpc_v0.NewJsonRpcServer(tmjs)
	restServer := rpc_v0.NewRestServer(codec, core.pipe, eventSubscriptions)
	wsServer := server.NewWebSocketServer(config.WebSocket.MaxWebSocketSessions,
		tmwss, core.logger)
	// Create a server process.
	proc, err := server.NewServeProcess(config, core.logger, jsonServer, restServer, wsServer)
	if err != nil {
		return nil, fmt.Errorf("Failed to load gateway: %v", err)
	}

	return proc, nil
}

func (core *Core) NewGatewayTendermint(config *server.ServerConfig) (
	*rpc_tendermint.TendermintWebsocketServer, error) {
	return rpc_tendermint.NewTendermintWebsocketServer(config,
		core.tendermintPipe, core.eventSwitch)
}

// Stop the core allowing for a graceful shutdown of component in order.
func (core *Core) Stop() bool {
	return core.pipe.GetConsensusEngine().Stop()
}
