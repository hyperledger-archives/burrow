// Starts a server using a mock pipe for manual testing.
package main

import (
	"github.com/eris-ltd/erisdb/server"
	edb "github.com/eris-ltd/erisdb/erisdb"
	"github.com/eris-ltd/erisdb/test/mock"
)

func main() {
	mockData := mock.NewDefaultMockData()
	mockPipe := mock.NewMockPipe(mockData)
	codec := &edb.TCodec{}
	tmwss := edb.NewErisDbWsService(codec, mockPipe)
	tmjs := edb.NewErisDbJsonService(codec, mockPipe, nil)
	
	// The servers.
	jsonServer := edb.NewJsonRpcServer(tmjs)
	restServer := edb.NewRestServer(codec, mockPipe, nil)
	wsServer := server.NewWebSocketServer(100, tmwss)
	proc := server.NewServeProcess(nil, jsonServer, restServer, wsServer)
	err := proc.Start()
	if err != nil {
		panic(err.Error())
	}
	<- proc.StopEventChannel()
}