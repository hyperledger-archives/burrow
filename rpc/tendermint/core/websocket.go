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
	"net"
	"net/http"
	"strings"

	events "github.com/tendermint/go-events"
	rpcserver "github.com/tendermint/go-rpc/server"

	definitions "github.com/hyperledger/burrow/definitions"
	server "github.com/hyperledger/burrow/server"
)

type TendermintWebsocketServer struct {
	routes    TendermintRoutes
	listeners []net.Listener
}

func NewTendermintWebsocketServer(config *server.ServerConfig,
	tendermintPipe definitions.TendermintPipe, evsw events.EventSwitch) (
	*TendermintWebsocketServer, error) {

	if tendermintPipe == nil {
		return nil, fmt.Errorf("No Tendermint pipe provided.")
	}
	tendermintRoutes := TendermintRoutes{
		tendermintPipe: tendermintPipe,
	}
	routes := tendermintRoutes.GetRoutes()
	listenerAddresses := strings.Split(config.Tendermint.RpcLocalAddress, ",")
	if len(listenerAddresses) == 0 {
		return nil, fmt.Errorf("No RPC listening addresses provided in [servers.tendermint.rpc_local_address] in configuration file: %s",
			listenerAddresses)
	}
	listeners := make([]net.Listener, len(listenerAddresses))
	for i, listenerAddress := range listenerAddresses {
		mux := http.NewServeMux()
		wm := rpcserver.NewWebsocketManager(routes, evsw)
		mux.HandleFunc(config.Tendermint.Endpoint, wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(mux, routes)
		listener, err := rpcserver.StartHTTPServer(listenerAddress, mux)
		if err != nil {
			return nil, err
		}
		listeners[i] = listener
	}
	return &TendermintWebsocketServer{
		routes:    tendermintRoutes,
		listeners: listeners,
	}, nil
}

func (tmServer *TendermintWebsocketServer) Shutdown() {
	for _, listener := range tmServer.listeners {
		listener.Close()
	}
}
