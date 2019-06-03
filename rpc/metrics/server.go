// Copyright 2019 Monax Industries Limited
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
package metrics

import (
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/lib/server"
)

func StartServer(service *rpc.Service, pattern string, listener net.Listener, blockSampleSize int,
	logger *logging.Logger) (*http.Server, error) {

	// instantiate metrics and variables we do not expect to change during runtime
	exporter, err := NewExporter(service, blockSampleSize, logger)
	if err != nil {
		return nil, err
	}

	// Register Metrics from each of the endpoints
	// This invokes the Collect method through the prometheus client libraries.
	prometheus.MustRegister(exporter)

	mux := http.NewServeMux()
	mux.Handle(pattern, server.RecoverAndLogHandler(promhttp.Handler(), logger))

	srv, err := server.StartHTTPServer(listener, mux, logger)
	if err != nil {
		return nil, err
	}
	return srv, nil
}
