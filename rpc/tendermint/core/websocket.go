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
	"net"

	rpcserver "github.com/tendermint/go-rpc"
	events    "github.com/tendermint/go-events"

	definitions "github.com/eris-ltd/eris-db/definitions"
)

type TendermintWebsocketServer struct {
	tendermintPipe definitions.TendermintPipe
	listeners      []net.Listeners
}

func NewTendermintWebsocketServer(config *server.ServerConfig,
	tendermintPipe definitions.TendermintPipe, evsw *events.EventSwitch) (
	*TendermintWebsocketServer, error) {

	listenersAddresses := strings.Split(config.Tendermint.RpcLocalAddress, ",")
	listeners := make([]net.Listeners, len(listenersAddresses))
	for i, listenerAddress := range listenersAddresses {
		mux := http.NewServeMux()
		wm := rpcserver.NewWebsocketManager(Routes, evsw)
		mux.HandleFunc(config.Tendermint.Endpoint, wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(mux, Routes)
		listener, err := rpcserver.StartHTTPServer(listenerAddress, mux)
		if err != nil {
			return nil, err
		}
		listeners[i] = listener
	}
	return &TendermintWebsocketServer {
		tendermintPipe: tendermintPipe,
		listeners:      listeners,
	}, nil
}
