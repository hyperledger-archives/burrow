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

package rpcinfo

import (
	"net/http"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/lib/server"
)

func StartServer(service *rpc.Service, pattern, listenAddress string, logger *logging.Logger) (*http.Server, error) {
	logger = logger.With(structure.ComponentKey, "RPC_Info")
	routes := GetRoutes(service, logger)
	mux := http.NewServeMux()
	wm := server.NewWebsocketManager(routes, logger)
	mux.HandleFunc(pattern, wm.WebsocketHandler)
	server.RegisterRPCFuncs(mux, routes, logger)
	srv, err := server.StartHTTPServer(listenAddress, mux, logger)
	if err != nil {
		return nil, err
	}
	return srv, nil
}
