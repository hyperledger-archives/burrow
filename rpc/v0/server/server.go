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

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/burrow/logging"
	"github.com/tommy351/gin-cors"
	"gopkg.in/tylerb/graceful.v1"
)

var (
	killTime = 100 * time.Millisecond
)

type HttpService interface {
	Process(*http.Request, http.ResponseWriter)
}

// A server serves a number of different http calls.
type Server interface {
	Start(*ServerConfig, *gin.Engine)
	Running() bool
	Shutdown(ctx context.Context) error
}

// The ServeProcess wraps all the Servers. Starting it will
// add all the server handlers to the router and start listening
// for incoming requests. There is also startup and shutdown events
// that can be listened to, on top of any events that the servers
// may have (the default websocket server has events for monitoring
// sessions. Startup event listeners should be added before calling
// 'Start()'. Stop event listeners can be added up to the point where
// the server is stopped and the event is fired.
type ServeProcess struct {
	config           *ServerConfig
	servers          []Server
	stopChan         chan struct{}
	startListenChans []chan struct{}
	stopListenChans  []chan struct{}
	srv              *graceful.Server
	logger           *logging.Logger
}

// Initializes all the servers and starts listening for connections.
func (serveProcess *ServeProcess) Start() error {
	router := gin.New()
	gin.SetMode(gin.ReleaseMode)
	config := serveProcess.config

	ch := NewCORSMiddleware(config.CORS)
	router.Use(gin.Recovery(), logHandler(serveProcess.logger), contentTypeMW, ch)

	address := config.Bind.Address
	port := config.Bind.Port

	if port == 0 {
		return fmt.Errorf("0 is not a valid port.")
	}

	listenAddress := address + ":" + fmt.Sprintf("%d", port)
	srv := &graceful.Server{
		Server: &http.Server{
			Handler: router,
		},
	}

	// Start the servers/handlers.
	for _, s := range serveProcess.servers {
		s.Start(config, router)
	}

	var lst net.Listener
	l, lErr := net.Listen("tcp", listenAddress)
	if lErr != nil {
		return lErr
	}

	// For secure connections.
	if config.TLS.TLS {
		addr := srv.Addr
		if addr == "" {
			addr = ":https"
		}

		tConfig := &tls.Config{}
		if tConfig.NextProtos == nil {
			tConfig.NextProtos = []string{"http/1.1"}
		}

		var tErr error
		tConfig.Certificates = make([]tls.Certificate, 1)
		tConfig.Certificates[0], tErr = tls.LoadX509KeyPair(config.TLS.CertPath, config.TLS.KeyPath)
		if tErr != nil {
			return tErr
		}

		lst = tls.NewListener(l, tConfig)
	} else {
		lst = l
	}
	serveProcess.srv = srv
	serveProcess.logger.InfoMsg("Server started.",
		"chain_id", serveProcess.config.ChainId,
		"address", serveProcess.config.Bind.Address,
		"port", serveProcess.config.Bind.Port)
	for _, c := range serveProcess.startListenChans {
		c <- struct{}{}
	}
	// Start the serve routine.
	go func() {
		serveProcess.srv.Serve(lst)
		for _, s := range serveProcess.servers {
			s.Shutdown(context.Background())
		}
	}()
	// Listen to the process stop event, it will call 'Stop'
	// on the graceful Server. This happens when someone
	// calls 'Stop' on the process.
	go func() {
		<-serveProcess.stopChan
		serveProcess.logger.InfoMsg("Close signal sent to server.")
		serveProcess.srv.Stop(killTime)
	}()
	// Listen to the servers stop event. It is triggered when
	// the server has been fully shut down.
	go func() {
		<-serveProcess.srv.StopChan()
		serveProcess.logger.InfoMsg("Server stop event fired. Good bye.")
		for _, c := range serveProcess.stopListenChans {
			c <- struct{}{}
		}
	}()
	return nil
}

// Stop will release the port, process any remaining requests
// up until the timeout duration is passed, at which point it
// will abort them and shut down.
func (serveProcess *ServeProcess) Shutdown(ctx context.Context) error {
	var err error
	for _, s := range serveProcess.servers {
		serr := s.Shutdown(ctx)
		if serr != nil && err == nil {
			err = serr
		}
	}

	lChan := serveProcess.StopEventChannel()
	serveProcess.stopChan <- struct{}{}
	select {
	case <-lChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Get a start-event channel from the server. The start event
// is fired after the Start() function is called, and after
// the server has started listening for incoming connections.
// An error here .
func (serveProcess *ServeProcess) StartEventChannel() <-chan struct{} {
	lChan := make(chan struct{}, 1)
	serveProcess.startListenChans = append(serveProcess.startListenChans, lChan)
	return lChan
}

// Get a stop-event channel from the server. The event happens
// after the Stop() function has been called, and after the
// timeout has passed. When the timeout has passed it will wait
// for confirmation from the http.Server, which normally takes
// a very short time (milliseconds).
func (serveProcess *ServeProcess) StopEventChannel() <-chan struct{} {
	lChan := make(chan struct{}, 1)
	serveProcess.stopListenChans = append(serveProcess.stopListenChans, lChan)
	return lChan
}

// Creates a new serve process.
func NewServeProcess(config *ServerConfig, logger *logging.Logger,
	servers ...Server) (*ServeProcess, error) {
	var scfg ServerConfig
	if config == nil {
		return nil, fmt.Errorf("Nil passed as server configuration")
	} else {
		scfg = *config
	}
	stopChan := make(chan struct{}, 1)
	startListeners := make([]chan struct{}, 0)
	stopListeners := make([]chan struct{}, 0)
	sp := &ServeProcess{
		config:           &scfg,
		servers:          servers,
		stopChan:         stopChan,
		startListenChans: startListeners,
		stopListenChans:  stopListeners,
		srv:              nil,
		logger:           logger.WithScope("ServeProcess"),
	}
	return sp, nil
}

// Used to enable log15 logging instead of the default Gin logging.
func logHandler(logger *logging.Logger) gin.HandlerFunc {
	logger = logger.WithScope("ginLogHandler")
	return func(c *gin.Context) {

		path := c.Request.URL.Path

		// Process request
		c.Next()

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.String()

		logger.Info.Log("client_ip", clientIP,
			"status_code", statusCode,
			"method", method,
			"path", path,
			"error", comment)
	}
}

func NewCORSMiddleware(options CORS) gin.HandlerFunc {
	o := cors.Options{
		AllowCredentials: options.AllowCredentials,
		AllowHeaders:     options.AllowHeaders,
		AllowMethods:     options.AllowMethods,
		AllowOrigins:     options.AllowOrigins,
		ExposeHeaders:    options.ExposeHeaders,
		MaxAge:           time.Duration(options.MaxAge),
	}
	return cors.Middleware(o)
}

// Just a catch-all for POST requests right now. Only allow default charset (utf8).
func contentTypeMW(c *gin.Context) {
	if c.Request.Method == "POST" && c.ContentType() != "application/json" {
		c.AbortWithError(415, fmt.Errorf("Media type not supported: "+c.ContentType()))
	} else {
		c.Next()
	}
}
