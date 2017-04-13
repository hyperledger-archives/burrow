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
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
)

// TODO too much fluff. Should probably phase gorilla out and move closer
// to net in connections/session management. At some point...

const (
	// Time allowed to write a message to the peer.
	writeWait = 0 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 0 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 0 * time.Second
	// Maximum message size allowed from a peer.
	maxMessageSize = 1000000
)

// Services requests. Message bytes are passed along with the session
// object. The service is expected to write any response back using
// the Write function on WSSession, which passes the message over
// a channel to the write pump.
type WebSocketService interface {
	Process([]byte, *WSSession)
}

// The websocket server handles incoming websocket connection requests,
// upgrading, reading, writing, and session management. Handling the
// actual requests is delegated to a websocket service.
type WebSocketServer struct {
	upgrader       websocket.Upgrader
	running        bool
	maxSessions    uint16
	sessionManager *SessionManager
	config         *ServerConfig
	allOrigins     bool
	logger         logging_types.InfoTraceLogger
}

// Create a new server.
// maxSessions is the maximum number of active websocket connections that is allowed.
// NOTE: This is not the total number of connections allowed - only those that are
// upgraded to websockets. Requesting a websocket connection will fail with a 503 if
// the server is at capacity.
func NewWebSocketServer(maxSessions uint16, service WebSocketService,
	logger logging_types.InfoTraceLogger) *WebSocketServer {
	return &WebSocketServer{
		maxSessions:    maxSessions,
		sessionManager: NewSessionManager(maxSessions, service, logger),
		logger:         logging.WithScope(logger, "WebSocketServer"),
	}
}

// Start the server. Adds the handler to the router and sets everything up.
func (wsServer *WebSocketServer) Start(config *ServerConfig, router *gin.Engine) {

	wsServer.config = config

	wsServer.upgrader = websocket.Upgrader{
		ReadBufferSize: int(config.WebSocket.ReadBufferSize),
		// TODO Will this be enough for massive "get blockchain" requests?
		WriteBufferSize: int(config.WebSocket.WriteBufferSize),
	}
	wsServer.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	router.GET(config.WebSocket.WebSocketEndpoint, wsServer.handleFunc)
	wsServer.running = true
}

// Is the server currently running.
func (wsServer *WebSocketServer) Running() bool {
	return wsServer.running
}

// Shut the server down.
func (wsServer *WebSocketServer) ShutDown() {
	wsServer.sessionManager.Shutdown()
	wsServer.running = false
}

// Get the session-manager.
func (wsServer *WebSocketServer) SessionManager() *SessionManager {
	return wsServer.sessionManager
}

// Handler for websocket requests.
func (wsServer *WebSocketServer) handleFunc(c *gin.Context) {
	r := c.Request
	w := c.Writer
	// Upgrade to websocket.
	wsConn, uErr := wsServer.upgrader.Upgrade(w, r, nil)

	if uErr != nil {
		errMsg := "Failed to upgrade to websocket connection"
		http.Error(w, fmt.Sprintf("%s: %s", errMsg, uErr.Error()), 400)
		logging.InfoMsg(wsServer.logger, errMsg, "error", uErr)
		return
	}

	session, cErr := wsServer.sessionManager.createSession(wsConn)

	if cErr != nil {
		errMsg := "Failed to establish websocket connection"
		http.Error(w, fmt.Sprintf("%s: %s", errMsg, cErr.Error()), 503)
		logging.InfoMsg(wsServer.logger, errMsg, "error", cErr)
		return
	}

	// Start the connection.
	logging.InfoMsg(wsServer.logger, "New websocket connection",
		"session_id", session.id)
	session.Open()
}

// Used to track sessions. Will notify when a session are opened
// and closed.
type SessionObserver interface {
	NotifyOpened(*WSSession)
	NotifyClosed(*WSSession)
}

// WSSession wraps a gorilla websocket.Conn, which in turn wraps a
// net.Conn object. Writing is done using the 'Write([]byte)' method,
// which passes the bytes on to the write pump over a channel.
type WSSession struct {
	sessionManager *SessionManager
	id             uint
	wsConn         *websocket.Conn
	writeChan      chan []byte
	writeCloseChan chan struct{}
	service        WebSocketService
	opened         bool
	closed         bool
	logger         logging_types.InfoTraceLogger
}

// Write a text message to the client.
func (wsSession *WSSession) Write(msg []byte) error {
	if wsSession.closed {
		logging.InfoMsg(wsSession.logger, "Attempting to write to closed session.")
		return fmt.Errorf("Session is closed")
	}
	wsSession.writeChan <- msg
	return nil
}

// Private. Helper for writing control messages.
func (wsSession *WSSession) write(mt int, payload []byte) error {
	wsSession.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	return wsSession.wsConn.WriteMessage(mt, payload)
}

// Get the session id number.
func (wsSession *WSSession) Id() uint {
	return wsSession.id
}

// Starts the read and write pumps. Blocks on the former.
// Notifies all the observers.
func (wsSession *WSSession) Open() {
	wsSession.opened = true
	wsSession.sessionManager.notifyOpened(wsSession)
	go wsSession.writePump()
	wsSession.readPump()
}

// Closes the net connection and cleans up. Notifies all the observers.
func (wsSession *WSSession) Close() {
	if !wsSession.closed {
		wsSession.closed = true
		wsSession.wsConn.Close()
		wsSession.sessionManager.removeSession(wsSession.id)
		logging.InfoMsg(wsSession.logger, "Closing websocket connection.",
			"remaining_active_sessions", len(wsSession.sessionManager.activeSessions))
		wsSession.sessionManager.notifyClosed(wsSession)
	}
}

// Has the session been opened?
func (wsSession *WSSession) Opened() bool {
	return wsSession.opened
}

// Has the session been closed?
func (wsSession *WSSession) Closed() bool {
	return wsSession.closed
}

// Pump debugging
/*
var rp int = 0
var wp int = 0
var rpm *sync.Mutex = &sync.Mutex{}
var wpm *sync.Mutex = &sync.Mutex{}
*/

// Read loop. Will terminate on a failed read.
func (wsSession *WSSession) readPump() {
	/*
		rpm.Lock()
		rp++
		log.Debug("readpump created", "total", rp)
		rpm.Unlock()
		defer func(){
			rpm.Lock()
			rp--
			log.Debug("readpump removed", "total", rp)
			rpm.Unlock()
			}()
	*/
	wsSession.wsConn.SetReadLimit(maxMessageSize)
	// this.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	// this.wsConn.SetPongHandler(func(string) error { this.wsConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		// Read
		msgType, msg, err := wsSession.wsConn.ReadMessage()

		// Read error.
		if err != nil {
			// Socket could have been gracefully closed, so not really an error.
			logging.InfoMsg(wsSession.logger,
				"Socket closed. Removing.", "error", err)
			wsSession.writeCloseChan <- struct{}{}
			return
		}

		if msgType != websocket.TextMessage {
			logging.InfoMsg(wsSession.logger,
				"Receiving non text-message from client, closing.")
			wsSession.writeCloseChan <- struct{}{}
			return
		}

		go func() {
			// Process the request.
			wsSession.service.Process(msg, wsSession)
		}()
	}
}

// Writes messages coming in on the write channel. Will terminate on failed writes,
// if pings are not responded to, or if a message comes in on the write close channel.
func (wsSession *WSSession) writePump() {
	/*
		wpm.Lock()
		wp++
		log.Debug("writepump created", "total", wp)
		wpm.Unlock()
		defer func() {
			wpm.Lock()
			wp--
			log.Debug("writepump removed", "total", wp)
			wpm.Unlock()
		}()
	*/
	// ticker := time.NewTicker(pingPeriod)

	defer func() {
		// ticker.Stop()
		wsSession.Close()
	}()

	// Write loop. Blocks while waiting for data to come in over a channel.
	for {
		select {
		// Write request.
		case msg := <-wsSession.writeChan:

			// Write the bytes to the socket.
			err := wsSession.wsConn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				// Could be due to the socket being closed so not really an error.
				logging.InfoMsg(wsSession.logger,
					"Writing to socket failed. Closing.")
				return
			}
		case <-wsSession.writeCloseChan:
			return
			// Ticker run out. Time for another ping message.
			/*
				case <-ticker.C:
					if err := this.write(websocket.PingMessage, []byte{}); err != nil {
						log.Debug("Failed to write ping message to socket. Closing.")
						return
					}
			*/
		}

	}
}

// Session manager handles the adding, tracking and removing of session objects.
type SessionManager struct {
	maxSessions     uint16
	activeSessions  map[uint]*WSSession
	idPool          *IdPool
	mtx             *sync.Mutex
	service         WebSocketService
	openEventChans  []chan *WSSession
	closeEventChans []chan *WSSession
	logger          logging_types.InfoTraceLogger
}

// Create a new WebsocketManager.
func NewSessionManager(maxSessions uint16, wss WebSocketService,
	logger logging_types.InfoTraceLogger) *SessionManager {
	return &SessionManager{
		maxSessions:     maxSessions,
		activeSessions:  make(map[uint]*WSSession),
		idPool:          NewIdPool(uint(maxSessions)),
		mtx:             &sync.Mutex{},
		service:         wss,
		openEventChans:  []chan *WSSession{},
		closeEventChans: []chan *WSSession{},
		logger:          logging.WithScope(logger, "SessionManager"),
	}
}

// TODO
func (sessionManager *SessionManager) Shutdown() {
	sessionManager.activeSessions = nil
}

// Add a listener to session open events.
func (sessionManager *SessionManager) SessionOpenEventChannel() <-chan *WSSession {
	lChan := make(chan *WSSession, 1)
	sessionManager.openEventChans = append(sessionManager.openEventChans, lChan)
	return lChan
}

// Remove a listener from session open events.
func (sessionManager *SessionManager) RemoveSessionOpenEventChannel(lChan chan *WSSession) bool {
	ec := sessionManager.openEventChans
	if len(ec) == 0 {
		return false
	}
	for i, c := range ec {
		if lChan == c {
			ec[i], ec = ec[len(ec)-1], ec[:len(ec)-1]
			return true
		}
	}
	return false
}

// Add a listener to session close events
func (sessionManager *SessionManager) SessionCloseEventChannel() <-chan *WSSession {
	lChan := make(chan *WSSession, 1)
	sessionManager.closeEventChans = append(sessionManager.closeEventChans, lChan)
	return lChan
}

// Remove a listener from session close events.
func (sessionManager *SessionManager) RemoveSessionCloseEventChannel(lChan chan *WSSession) bool {
	ec := sessionManager.closeEventChans
	if len(ec) == 0 {
		return false
	}
	for i, c := range ec {
		if lChan == c {
			ec[i], ec = ec[len(ec)-1], ec[:len(ec)-1]
			return true
		}
	}
	return false
}

// Used to notify all observers that a new session was opened.
func (sessionManager *SessionManager) notifyOpened(session *WSSession) {
	for _, lChan := range sessionManager.openEventChans {
		lChan <- session
	}
}

// Used to notify all observers that a new session was closed.
func (sessionManager *SessionManager) notifyClosed(session *WSSession) {
	for _, lChan := range sessionManager.closeEventChans {
		lChan <- session
	}
}

// Creates a new session and adds it to the manager.
func (sessionManager *SessionManager) createSession(wsConn *websocket.Conn) (*WSSession, error) {
	// Check that the capacity hasn't been exceeded.
	sessionManager.mtx.Lock()
	defer sessionManager.mtx.Unlock()
	if sessionManager.atCapacity() {
		return nil, fmt.Errorf("Already at capacity")
	}

	// Create and start
	newId, _ := sessionManager.idPool.GetId()
	conn := &WSSession{
		sessionManager: sessionManager,
		id:             newId,
		wsConn:         wsConn,
		writeChan:      make(chan []byte, maxMessageSize),
		writeCloseChan: make(chan struct{}),
		service:        sessionManager.service,
		logger: logging.WithScope(sessionManager.logger, "WSSession").
			With("session_id", newId),
	}
	sessionManager.activeSessions[conn.id] = conn
	return conn, nil
}

// Remove a session from the list.
func (sessionManager *SessionManager) removeSession(id uint) {
	sessionManager.mtx.Lock()
	defer sessionManager.mtx.Unlock()
	// Check that it exists.
	_, ok := sessionManager.activeSessions[id]
	if ok {
		delete(sessionManager.activeSessions, id)
		sessionManager.idPool.ReleaseId(id)
	}
}

// True if the number of active connections is at the maximum.
func (sessionManager *SessionManager) atCapacity() bool {
	return len(sessionManager.activeSessions) >= int(sessionManager.maxSessions)
}
