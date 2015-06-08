package test

import (
	"fmt"
	"github.com/eris-ltd/erisdb/server"
	"github.com/eris-ltd/erisdb/rpc"
	"github.com/gin-gonic/gin"
	"encoding/json"
)

type ScumbagServer struct {
	running bool
}

func NewScumbagServer() server.Server {
	return &ScumbagServer{}
}

func (this *ScumbagServer) Start(sc *server.ServerConfig, g *gin.Engine) {
	g.GET("/scumbag", func(c *gin.Context){
				c.String(200, "Scumbag")
			})
	this.running = true
}

func (this *ScumbagServer) Running() bool {
	return this.running;
}

func (this *ScumbagServer) ShutDown() {
	fmt.Println("Scumbag...")
}

type ScumSocketService struct {}

func (this *ScumSocketService) Process(data []byte, session *server.WSSession){
	resp := rpc.NewRPCResponse("1", "Scumbag")
	bts, _ := json.Marshal(resp)
	session.Write(bts)
}

func NewScumsocketServer(maxConnections uint) *server.WebSocketServer {
	sss := &ScumSocketService{}
	return server.NewWebSocketServer(maxConnections, sss)
}

func NewServeScumbag() *server.ServeProcess {
	return server.NewServeProcess(nil, NewScumbagServer())
}

func NewServeScumSocket(wsServer *server.WebSocketServer) *server.ServeProcess{
	cfg := server.DefaultServerConfig()
	cfg.WebSocketPath = "/scumsocket"
	cfg.Port = uint16(31337)
	return server.NewServeProcess(cfg, wsServer)	
}