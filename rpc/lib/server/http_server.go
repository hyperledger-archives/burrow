// Commons for HTTP handling
package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/rpc/lib/types"
)

func StartHTTPServer(listener net.Listener, handler http.Handler, logger *logging.Logger) (*http.Server, error) {
	logger.InfoMsg("Starting RPC HTTP server", "listen_address", listener.Addr().String())

	server := &http.Server{Handler: RecoverAndLogHandler(handler, logger)}

	go func() {
		err := server.Serve(listener)
		logger.TraceMsg("RPC HTTP server stopped", structure.ErrorKey, err)
	}()

	return server, nil
}

func WriteRPCResponseHTTP(w http.ResponseWriter, res types.RPCResponse) {
	jsonBytes, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.Error.HTTPStatusCode())
	w.Write(jsonBytes) // nolint: errcheck, gas
}

//-----------------------------------------------------------------------------

// Wraps an HTTP handler, adding error logging.
// If the inner function panics, the outer function recovers, logs, sends an
// HTTP 500 error response.
func RecoverAndLogHandler(handler http.Handler, logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the ResponseWriter to remember the status
		rww := &ResponseWriterWrapper{-1, w}
		begin := time.Now()

		// Common headers
		origin := r.Header.Get("Origin")
		rww.Header().Set("Access-Control-Allow-Origin", origin)
		rww.Header().Set("Access-Control-Allow-Credentials", "true")
		rww.Header().Set("Access-Control-Expose-Headers", "X-Server-Time")
		rww.Header().Set("X-Server-Time", fmt.Sprintf("%v", begin.Unix()))

		defer func() {
			// Send a 500 error if a panic happens during a handler.
			// Without this, Chrome & Firefox were retrying aborted ajax requests,
			// at least to my localhost.
			if e := recover(); e != nil {

				// If RPCResponse
				if res, ok := e.(types.RPCResponse); ok {
					WriteRPCResponseHTTP(rww, res)
				} else {
					// For the rest,
					logger.TraceMsg("Panic in RPC HTTP handler", structure.ErrorKey, e,
						"stack", string(debug.Stack()))
					rww.WriteHeader(http.StatusInternalServerError)
					WriteRPCResponseHTTP(rww, types.RPCInternalError("", e.(error)))
				}
			}

			// Finally, log.
			duration := time.Since(begin)
			if rww.Status == -1 {
				rww.Status = 200
			}
			logger.InfoMsg("Served RPC HTTP response",
				"method", r.Method,
				"url", r.URL,
				"status", rww.Status,
				"duration", duration,
				"remote_address", r.RemoteAddr,
			)
		}()

		handler.ServeHTTP(rww, r)
	})
}

// Remember the status for logging
type ResponseWriterWrapper struct {
	Status int
	http.ResponseWriter
}

func (w *ResponseWriterWrapper) WriteHeader(status int) {
	w.Status = status
	w.ResponseWriter.WriteHeader(status)
}

// implements http.Hijacker
func (w *ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
