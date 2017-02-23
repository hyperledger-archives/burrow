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

package tmsp

// Taken originally from github.com/tendermint/tmsp/server.go

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	. "github.com/tendermint/go-common"
	tmsp_types "github.com/tendermint/tmsp/types"

	manager_types "github.com/eris-ltd/eris-db/manager/types"
)

// var maxNumberConnections = 2

type Server struct {
	QuitService

	proto    string
	addr     string
	listener net.Listener

	appMtx sync.Mutex
	app    manager_types.Application
}

func NewServer(protoAddr string, app manager_types.Application) (*Server, error) {
	parts := strings.SplitN(protoAddr, "://", 2)
	proto, addr := parts[0], parts[1]
	s := &Server{
		proto:    proto,
		addr:     addr,
		listener: nil,
		app:      app,
	}
	s.QuitService = *NewQuitService(nil, "TMSPServer", s)
	_, err := s.Start() // Just start it
	return s, err
}

func (s *Server) OnStart() error {
	s.QuitService.OnStart()
	ln, err := net.Listen(s.proto, s.addr)
	if err != nil {
		return err
	}
	s.listener = ln
	go s.acceptConnectionsRoutine()
	return nil
}

func (s *Server) OnStop() {
	s.QuitService.OnStop()
	s.listener.Close()
}

func (s *Server) acceptConnectionsRoutine() {
	// semaphore := make(chan struct{}, maxNumberConnections)

	for {
		// semaphore <- struct{}{}

		// Accept a connection
		fmt.Println("Waiting for new connection...")
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.IsRunning() {
				return // Ignore error from listener closing.
			}
			Exit("Failed to accept connection: " + err.Error())
		} else {
			fmt.Println("Accepted a new connection")
		}

		closeConn := make(chan error, 2)                   // Push to signal connection closed
		responses := make(chan *tmsp_types.Response, 1000) // A channel to buffer responses

		// Read requests from conn and deal with them
		go s.handleRequests(closeConn, conn, responses)
		// Pull responses from 'responses' and write them to conn.
		go s.handleResponses(closeConn, responses, conn)

		go func() {
			// Wait until signal to close connection
			errClose := <-closeConn
			if errClose != nil {
				fmt.Printf("Connection error: %v\n", errClose)
			} else {
				fmt.Println("Connection was closed.")
			}

			// Close the connection
			err := conn.Close()
			if err != nil {
				fmt.Printf("Error in closing connection: %v\n", err)
			}

			// <-semaphore
		}()
	}
}

// Read requests from conn and deal with them
func (s *Server) handleRequests(closeConn chan error, conn net.Conn, responses chan<- *tmsp_types.Response) {
	var count int
	var bufReader = bufio.NewReader(conn)
	for {

		var req = &tmsp_types.Request{}
		err := tmsp_types.ReadMessage(bufReader, req)
		if err != nil {
			if err == io.EOF {
				closeConn <- fmt.Errorf("Connection closed by client")
			} else {
				closeConn <- fmt.Errorf("Error in handleRequests: %v", err.Error())
			}
			return
		}
		s.appMtx.Lock()
		count++
		s.handleRequest(req, responses)
		s.appMtx.Unlock()
	}
}

func (s *Server) handleRequest(req *tmsp_types.Request, responses chan<- *tmsp_types.Response) {
	switch r := req.Value.(type) {
	case *tmsp_types.Request_Echo:
		responses <- tmsp_types.ToResponseEcho(r.Echo.Message)
	case *tmsp_types.Request_Flush:
		responses <- tmsp_types.ToResponseFlush()
	case *tmsp_types.Request_Info:
		data := s.app.Info()
		responses <- tmsp_types.ToResponseInfo(data)
	case *tmsp_types.Request_SetOption:
		so := r.SetOption
		logStr := s.app.SetOption(so.Key, so.Value)
		responses <- tmsp_types.ToResponseSetOption(logStr)
	case *tmsp_types.Request_AppendTx:
		res := s.app.AppendTx(r.AppendTx.Tx)
		responses <- tmsp_types.ToResponseAppendTx(res.Code, res.Data, res.Log)
	case *tmsp_types.Request_CheckTx:
		res := s.app.CheckTx(r.CheckTx.Tx)
		responses <- tmsp_types.ToResponseCheckTx(res.Code, res.Data, res.Log)
	case *tmsp_types.Request_Commit:
		res := s.app.Commit()
		responses <- tmsp_types.ToResponseCommit(res.Code, res.Data, res.Log)
	case *tmsp_types.Request_Query:
		res := s.app.Query(r.Query.Query)
		responses <- tmsp_types.ToResponseQuery(res.Code, res.Data, res.Log)
	case *tmsp_types.Request_InitChain:
		if app, ok := s.app.(tmsp_types.BlockchainAware); ok {
			app.InitChain(r.InitChain.Validators)
			responses <- tmsp_types.ToResponseInitChain()
		} else {
			responses <- tmsp_types.ToResponseInitChain()
		}
	case *tmsp_types.Request_EndBlock:
		if app, ok := s.app.(tmsp_types.BlockchainAware); ok {
			validators := app.EndBlock(r.EndBlock.Height)
			responses <- tmsp_types.ToResponseEndBlock(validators)
		} else {
			responses <- tmsp_types.ToResponseEndBlock(nil)
		}
	default:
		responses <- tmsp_types.ToResponseException("Unknown request")
	}
}

// Pull responses from 'responses' and write them to conn.
func (s *Server) handleResponses(closeConn chan error, responses <-chan *tmsp_types.Response, conn net.Conn) {
	var count int
	var bufWriter = bufio.NewWriter(conn)
	for {
		var res = <-responses
		err := tmsp_types.WriteMessage(res, bufWriter)
		if err != nil {
			closeConn <- fmt.Errorf("Error in handleResponses: %v", err.Error())
			return
		}
		if _, ok := res.Value.(*tmsp_types.Response_Flush); ok {
			err = bufWriter.Flush()
			if err != nil {
				closeConn <- fmt.Errorf("Error in handleValue: %v", err.Error())
				return
			}
		}
		count++
	}
}
