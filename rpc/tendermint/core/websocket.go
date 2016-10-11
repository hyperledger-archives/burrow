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
	"net"
	"net/http"
	"strings"

	events "github.com/tendermint/go-events"
	rpcserver "github.com/tendermint/go-rpc/server"

	definitions "github.com/eris-ltd/eris-db/definitions"
	server "github.com/eris-ltd/eris-db/server"
)

type TendermintWebsocketServer struct {
	routes    TendermintRoutes
	listeners []net.Listener
}

func NewTendermintWebsocketServer(config *server.ServerConfig,
	tendermintPipe definitions.TendermintPipe, evsw *events.EventSwitch) (
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
