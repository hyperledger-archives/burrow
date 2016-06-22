// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"fmt"
	"net/http"

	// TODO: [ben] swap out go-events with eris-db/event (currently unused)
	events "github.com/tendermint/go-events"

	log "github.com/eris-ltd/eris-logger"

	config "github.com/eris-ltd/eris-db/config"
	consensus "github.com/eris-ltd/eris-db/consensus"
	definitions "github.com/eris-ltd/eris-db/definitions"
	event "github.com/eris-ltd/eris-db/event"
	manager "github.com/eris-ltd/eris-db/manager"
	erismint "github.com/eris-ltd/eris-db/manager/eris-mint"
	rpccore "github.com/eris-ltd/eris-db/rpc/tendermint/core"
	// rpc_v0 is carried over from Eris-DBv0.11 and before on port 1337
	rpc_v0 "github.com/eris-ltd/eris-db/rpc/v0"
	// rpc_tendermint is carried over from Eris-DBv0.11 and before on port 46657
	// rpc_tendermint "github.com/eris-ltd/eris-db/rpc/tendermint"
	server "github.com/eris-ltd/eris-db/server"
	"github.com/tendermint/go-rpc/server"
	"github.com/tendermint/tendermint/node"
	"net"
	"strings"
)

// Core is the high-level structure
type Core struct {
	chainId string
	evsw    *events.EventSwitch
	pipe    definitions.Pipe
}

func NewCore(chainId string, consensusConfig *config.ModuleConfig,
	managerConfig *config.ModuleConfig) (*Core, error) {
	// start new event switch, TODO: [ben] replace with eris-db/event
	evsw := events.NewEventSwitch()
	evsw.Start()

	// start a new application pipe that will load an application manager
	pipe, err := manager.NewApplicationPipe(managerConfig, evsw,
		consensusConfig.Version)
	if err != nil {
		return nil, fmt.Errorf("Failed to load application pipe: %v", err)
	}
	log.Debug("Loaded pipe with application manager")
	// pass the consensus engine into the pipe
	consensus.LoadConsensusEngineInPipe(consensusConfig, pipe)

	return &Core{
		chainId: chainId,
		evsw:    evsw,
		pipe:    pipe,
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

func StartRPC(config config.ModuleConfig, n *node.Node, erisMint *erismint.ErisMint) ([]net.Listener, error) {
	rpccore.SetConfig(config)

	rpccore.SetErisMint(erisMint)
	rpccore.SetBlockStore(n.BlockStore())
	rpccore.SetConsensusState(n.ConsensusState())
	rpccore.SetConsensusReactor(n.ConsensusReactor())
	rpccore.SetMempoolReactor(n.MempoolReactor())
	rpccore.SetSwitch(n.Switch())
	rpccore.SetPrivValidator(n.PrivValidator())
	// TODO: programming
	//rpccore.SetGenDoc(LoadGenDoc(config.GetString("genesis_file")))

	// TODO: also programming
	//listenAddrs := strings.Split(config.GetString("rpc_laddr"), ",")
	listenAddrs := strings.Split("127.0.0.1", ",")

	// we may expose the rpc over both a unix and tcp socket
	listeners := make([]net.Listener, len(listenAddrs))
	for i, listenAddr := range listenAddrs {
		mux := http.NewServeMux()
		wm := rpcserver.NewWebsocketManager(rpccore.Routes, n.EventSwitch())
		mux.HandleFunc("/websocket", wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(mux, rpccore.Routes)
		listener, err := rpcserver.StartHTTPServer(listenAddr, mux)
		if err != nil {
			return nil, err
		}
		listeners[i] = listener
	}
	return listeners, nil
}
