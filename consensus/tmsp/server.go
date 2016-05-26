// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

// Taken originally from github.com/tendermint/tmsp/server.go

package tmsp

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

		closeConn := make(chan error, 2)              // Push to signal connection closed
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
	switch req.Type {
	case tmsp_types.MessageType_Echo:
		responses <- tmsp_types.ResponseEcho(string(req.Data))
	case tmsp_types.MessageType_Flush:
		responses <- tmsp_types.ResponseFlush()
	case tmsp_types.MessageType_Info:
		data := s.app.Info()
		responses <- tmsp_types.ResponseInfo(data)
	case tmsp_types.MessageType_SetOption:
		logStr := s.app.SetOption(req.Key, req.Value)
		responses <- tmsp_types.ResponseSetOption(logStr)
	case tmsp_types.MessageType_AppendTx:
		res := s.app.AppendTx(req.Data)
		responses <- tmsp_types.ResponseAppendTx(res.Code, res.Data, res.Log)
	case tmsp_types.MessageType_CheckTx:
		res := s.app.CheckTx(req.Data)
		responses <- tmsp_types.ResponseCheckTx(res.Code, res.Data, res.Log)
	case tmsp_types.MessageType_Commit:
		res := s.app.Commit()
		responses <- tmsp_types.ResponseCommit(res.Code, res.Data, res.Log)
	case tmsp_types.MessageType_Query:
		res := s.app.Query(req.Data)
		responses <- tmsp_types.ResponseQuery(res.Code, res.Data, res.Log)
	case tmsp_types.MessageType_InitChain:
		if app, ok := s.app.(manager_types.BlockchainAware); ok {
			app.InitChain(req.Validators)
			responses <- tmsp_types.ResponseInitChain()
		} else {
			responses <- tmsp_types.ResponseInitChain()
		}
	case tmsp_types.MessageType_EndBlock:
		if app, ok := s.app.(manager_types.BlockchainAware); ok {
			validators := app.EndBlock(req.Height)
			responses <- tmsp_types.ResponseEndBlock(validators)
		} else {
			responses <- tmsp_types.ResponseEndBlock(nil)
		}
	default:
		responses <- tmsp_types.ResponseException("Unknown request")
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
		if res.Type == tmsp_types.MessageType_Flush {
			err = bufWriter.Flush()
			if err != nil {
				closeConn <- fmt.Errorf("Error in handleResponses: %v", err.Error())
				return
			}
		}
		count++
	}
}
