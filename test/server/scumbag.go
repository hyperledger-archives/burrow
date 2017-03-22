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
	"encoding/json"
	"os"
	"runtime"

	rpc "github.com/monax/eris-db/rpc"
	"github.com/monax/eris-db/server"
	"github.com/gin-gonic/gin"
	"github.com/tendermint/log15"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log15.Root().SetHandler(log15.LvlFilterHandler(
		log15.LvlWarn,
		log15.StreamHandler(os.Stdout, log15.TerminalFormat()),
	))
	gin.SetMode(gin.ReleaseMode)
}

type ScumbagServer struct {
	running bool
}

func NewScumbagServer() server.Server {
	return &ScumbagServer{}
}

func (this *ScumbagServer) Start(sc *server.ServerConfig, g *gin.Engine) {
	g.GET("/scumbag", func(c *gin.Context) {
		c.String(200, "Scumbag")
	})
	this.running = true
}

func (this *ScumbagServer) Running() bool {
	return this.running
}

func (this *ScumbagServer) ShutDown() {
	// fmt.Println("Scumbag...")
}

type ScumSocketService struct{}

func (this *ScumSocketService) Process(data []byte, session *server.WSSession) {
	resp := rpc.NewRPCResponse("1", "Scumbag")
	bts, _ := json.Marshal(resp)
	session.Write(bts)
}

func NewScumsocketServer(maxConnections uint16) *server.WebSocketServer {
	sss := &ScumSocketService{}
	return server.NewWebSocketServer(maxConnections, sss)
}

func NewServeScumbag() (*server.ServeProcess, error) {
	cfg := server.DefaultServerConfig()
	cfg.Bind.Port = uint16(31400)
	return server.NewServeProcess(cfg, NewScumbagServer())
}

func NewServeScumSocket(wsServer *server.WebSocketServer) (*server.ServeProcess,
	error) {
	cfg := server.DefaultServerConfig()
	cfg.WebSocket.WebSocketEndpoint = "/scumsocket"
	cfg.Bind.Port = uint16(31401)
	return server.NewServeProcess(cfg, wsServer)
}
