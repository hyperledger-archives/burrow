package server

import (
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	// cors "github.com/tommy351/gin-cors"
	"gopkg.in/tylerb/graceful.v1"
	"net"
	"net/http"
	"time"
)

var (
	killTime = 100 * time.Millisecond
)

// TODO should this be here.
type HttpService interface {
	Process(*http.Request, http.ResponseWriter)
}

// A server serves a number of different http calls.
type Server interface {
	Start(*ServerConfig, *gin.Engine)
	Running() bool
	ShutDown()
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
	stoppedChan      chan struct{}
	startListenChans []chan struct{}
	stopListenChans  []chan struct{}
	srv              *graceful.Server
}

// Initializes all the servers and starts listening for connections.
func (this *ServeProcess) Start() error {

	router := gin.New()

	config := this.config

	// ch := NewCORSMiddleware(config.CORS)
	// router.Use(gin.Recovery(), logHandler, ch)
	router.Use(gin.Recovery(), logHandler)

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
	for _, s := range this.servers {
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
	this.srv = srv
	log.Info("Server started.")
	for _, c := range this.startListenChans {
		c <- struct{}{}
	}
	go func() {
		this.srv.Serve(lst)
		for _, s := range this.servers {
			s.ShutDown()
		}
	}()
	go func() {
		<-this.stopChan
		log.Info("Close signal sent to server.")
		this.srv.Stop(killTime)
	}()
	go func() {
		<-this.srv.StopChan()
		log.Info("Server stop event fired. Good bye.")
		for _, c := range this.stopListenChans {
			c <- struct{}{}
		}
	}()
	return nil
}

// Stop will release the port, process any remaining requests
// up until the timeout duration is passed, at which point it
// will abort them and shut down.
func (this *ServeProcess) Stop(timeout time.Duration) error {
	toChan := make(chan struct{})
	if timeout != 0 {
		go func() {
			time.Sleep(timeout)
			toChan <- struct{}{}
		}()
	}

	lChan := this.StopEventChannel()
	this.stopChan <- struct{}{}
	select {
	case <-lChan:
		return nil
	case <-toChan:
		return fmt.Errorf("Timeout when stopping server")
	}
}

// Get a start-event channel from the server. The start event
// is fired after the Start() function is called, and after
// the server has started listening for incoming connections.
// An error here .
func (this *ServeProcess) StartEventChannel() <-chan struct{} {
	lChan := make(chan struct{}, 1)
	this.startListenChans = append(this.startListenChans, lChan)
	return lChan
}

// Get a stop-event channel from the server. The event happens
// after the Stop() function has been called, and after the
// timeout has passed. When the timeout has passed it will wait
// for confirmation from the http.Server, which normally takes
// a very short time (milliseconds).
func (this *ServeProcess) StopEventChannel() <-chan struct{} {
	lChan := make(chan struct{}, 1)
	this.stopListenChans = append(this.stopListenChans, lChan)
	return lChan
}

// Creates a new serve process.
func NewServeProcess(config *ServerConfig, servers ...Server) *ServeProcess {
	var cfg *ServerConfig
	if config == nil {
		cfg = DefaultServerConfig()
	} else {
		cfg = config
	}
	stopChan := make(chan struct{}, 1)
	stoppedChan := make(chan struct{}, 1)
	startListeners := make([]chan struct{}, 0)
	stopListeners := make([]chan struct{}, 0)
	sp := &ServeProcess{cfg, servers, stopChan, stoppedChan, startListeners, stopListeners, nil}
	return sp
}

// Used to enable log15 logging instead of the default Gin logging.
// This is done mainly because we at Eris uses log15 in other components.
// TODO make this optional perhaps.
func logHandler(c *gin.Context) {

	path := c.Request.URL.Path

	// Process request
	c.Next()

	clientIP := c.ClientIP()
	method := c.Request.Method
	statusCode := c.Writer.Status()
	comment := c.Errors.String()

	log.Info("[GIN] HTTP: "+clientIP, "Code", statusCode, "Method", method, "path", path, "error", comment)

}

/*
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
*/
