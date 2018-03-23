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

package tm

import (
	"net"
	"net/http"

	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/logging/structure"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/rpc"
	"github.com/tendermint/tendermint/rpc/lib/server"
)

func StartServer(service *rpc.Service, pattern, listenAddress string, emitter event.Emitter,
	logger logging_types.InfoTraceLogger) (net.Listener, error) {

	logger = logger.With(structure.ComponentKey, "RPC_TM")
	routes := GetRoutes(service, logger)
	mux := http.NewServeMux()
	wm := rpcserver.NewWebsocketManager(routes, rpcserver.EventSubscriber(tendermint.SubscribableAsEventBus(emitter)))
	mux.HandleFunc(pattern, wm.WebsocketHandler)
	tmLogger := tendermint.NewLogger(logger)
	rpcserver.RegisterRPCFuncs(mux, routes, tmLogger)
	listener, err := rpcserver.StartHTTPServer(listenAddress, mux, tmLogger)
	if err != nil {
		return nil, err
	}
	return listener, nil
}
