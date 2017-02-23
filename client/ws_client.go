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

package client

// NOTE: this websocket client acts on rpc/v0,
// uses github.com/gorilla/websocket
// and will be deprecated after 0.12
// It is recommended to use the interfaces NodeClient
// and NodeWebsocketClient.
// Websocket client implementation. This will be used in tests.

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// A websocket client subscribes and unsubscribes to events
type WSClient struct {
	host   string
	closed bool
	conn   *websocket.Conn
}

// create a new connection
func NewWSClient(addr string) *WSClient {
	return &WSClient{
		host: addr,
	}
}

func (this *WSClient) Dial() (*http.Response, error) {
	dialer := websocket.DefaultDialer
	rHeader := http.Header{}
	conn, r, err := dialer.Dial(this.host, rHeader)
	if err != nil {
		return r, err
	}
	this.conn = conn

	return r, nil
}

// returns a channel from which messages can be pulled
// from a go routine that reads the socket.
// if the ws returns an error (eg. closes), we return
func (this *WSClient) StartRead() <-chan []byte {
	ch := make(chan []byte)
	go func() {
		for {
			_, msg, err := this.conn.ReadMessage()
			if err != nil {
				if !this.closed {
					// TODO For now.
					fmt.Println("Error: " + err.Error())
					close(ch)
				}
				return
			}
			ch <- msg
		}
	}()
	return ch
}

func (this *WSClient) WriteMsg(msg []byte) {
	this.conn.WriteMessage(websocket.TextMessage, msg)
}

func (this *WSClient) Close() {
	this.closed = true
	this.conn.Close()
}
