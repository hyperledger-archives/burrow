package server

import (
	"fmt"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/gin-gonic/gin"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

// TODO too much fluff. Should probably phase gorilla out and move closer
// to net in connections/session management. At some point...

const (
	// Size of read channel.
	readChanBufferSize = 10
	// Size of write channel.
	writeChanBufferSize = 10
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from a peer.
	maxMessageSize = 2048
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
	maxSessions    uint
	sessionManager *SessionManager
	config         *ServerConfig
	allOrigins     bool
}

// Create a new server.
// maxSessions is the maximum number of active websocket connections that is allowed.
// NOTE: This is not the total number of connections allowed - only those that are
// upgraded to websockets. Requesting a websocket connection will fail with a 503 if
// the server is at capacity.
func NewWebSocketServer(maxSessions uint, service WebSocketService) *WebSocketServer {
	return &WebSocketServer{
		maxSessions:    maxSessions,
		sessionManager: NewSessionManager(maxSessions, service),
	}
}

// Start the server. Adds the handler to the router and sets everything up.
func (this *WebSocketServer) Start(config *ServerConfig, router *gin.Engine) {

	this.config = config

	this.upgrader = websocket.Upgrader{
		ReadBufferSize: int(config.WebSocket.ReadBufferSize),
		// TODO Will this be enough for massive "get blockchain" requests?
		WriteBufferSize: int(config.WebSocket.WriteBufferSize),
	}
	this.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	router.GET(config.WebSocket.WebSocketEndpoint, this.handleFunc)
	this.running = true
}

// Is the server currently running.
func (this *WebSocketServer) Running() bool {
	return this.running
}

// Shut the server down.
func (this *WebSocketServer) ShutDown() {
	this.sessionManager.Shutdown()
	this.running = false
}

// Get the session-manager.
func (this *WebSocketServer) SessionManager() *SessionManager {
	return this.sessionManager
}

// Handler for websocket requests.
func (this *WebSocketServer) handleFunc(c *gin.Context) {
	r := c.Request
	w := c.Writer
	// Upgrade to websocket.
	wsConn, uErr := this.upgrader.Upgrade(w, r, nil)

	if uErr != nil {
		uErrStr := "Failed to upgrade to websocket connection: " + uErr.Error()
		http.Error(w, uErrStr, 400)
		log.Info(uErrStr)
		return
	}

	session, cErr := this.sessionManager.createSession(wsConn)

	if cErr != nil {
		cErrStr := "Failed to establish websocket connection: " + cErr.Error()
		http.Error(w, cErrStr, 503)
		log.Info(cErrStr)
		return
	}

	// Start the connection.
	log.Info("New websocket connection.", "sessionId", session.id)
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
}

// Write a text message to the client.
func (this *WSSession) Write(msg []byte) error {
	if this.closed {
		log.Warn("Attempting to write to closed session.", "sessionId", this.id)
		return fmt.Errorf("Session is closed")
	}
	this.writeChan <- msg
	return nil
}

// Private. Helper for writing control messages.
func (this *WSSession) write(mt int, payload []byte) error {
	this.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	return this.wsConn.WriteMessage(mt, payload)
}

// Get the session id number.
func (this *WSSession) Id() uint {
	return this.id
}

// Starts the read and write pumps. Blocks on the former.
// Notifies all the observers.
func (this *WSSession) Open() {
	this.opened = true
	this.sessionManager.notifyOpened(this)
	go this.writePump()
	this.readPump()
}

// Closes the net connection and cleans up. Notifies all the observers.
func (this *WSSession) Close() {
	if !this.closed {
		this.closed = true
		this.wsConn.Close()
		this.sessionManager.removeSession(this.id)
		log.Debug("Closing websocket connection.", "sessionId", this.id, "remaining", len(this.sessionManager.activeSessions))
		this.sessionManager.notifyClosed(this)
	}
}

// Has the session been opened?
func (this *WSSession) Opened() bool {
	return this.opened
}

// Has the session been closed?
func (this *WSSession) Closed() bool {
	return this.closed
}

// Pump debugging
/*
var rp int = 0
var wp int = 0
var rpm *sync.Mutex = &sync.Mutex{}
var wpm *sync.Mutex = &sync.Mutex{}
*/

// Read loop. Will terminate on a failed read.
func (this *WSSession) readPump() {
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
	this.wsConn.SetReadLimit(maxMessageSize)
	this.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	this.wsConn.SetPongHandler(func(string) error { this.wsConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		// Read
		msgType, msg, err := this.wsConn.ReadMessage()

		// Read error.
		if err != nil {
			// Socket could have been gracefully closed, so not really an error.
			log.Debug("Socket closed. Removing.", "error", err.Error())
			this.writeCloseChan <- struct{}{}
			return
		}
		// Wrong message type.
		if msgType != websocket.TextMessage {
			var typeStr string
			if msgType == websocket.BinaryMessage {
				typeStr = "Binary"
			} else if msgType == websocket.CloseMessage {
				typeStr = "Close"
			} else if msgType == websocket.PingMessage {
				typeStr = "Ping"
			} else if msgType == websocket.PingMessage {
				typeStr = "Pong"
			} else {
				// This should not be possible.
				typeStr = "Unknown ID: " + fmt.Sprintf("%d", msgType)
			}

			log.Info("Receiving non text-message from client, closing.", "type", typeStr)
			this.writeCloseChan <- struct{}{}
			return
		}

		// Process the request.
		this.service.Process(msg, this)
	}
}

// Writes messages coming in on the write channel. Will terminate on failed writes,
// if pings are not responded to, or if a message comes in on the write close channel.
func (this *WSSession) writePump() {
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
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		this.Close()
	}()

	// Write loop. Blocks while waiting for data to come in over a channel.
	for {
		select {
		// Write request.
		case msg := <-this.writeChan:

			// Write the bytes to the socket.
			err := this.wsConn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				// Could be due to the socket being closed so not really an error.
				log.Info("Writing to socket failed. Closing.")
				return
			}
		case <-this.writeCloseChan:
			return
		// Ticker run out. Time for another ping message.
		case <-ticker.C:
			if err := this.write(websocket.PingMessage, []byte{}); err != nil {
				log.Debug("Failed to write ping message to socket. Closing.")
				return
			}
		}
	}
}

// Session manager handles the adding, tracking and removing of session objects.
type SessionManager struct {
	maxSessions     uint
	activeSessions  map[uint]*WSSession
	idPool          *IdPool
	mtx             *sync.Mutex
	service         WebSocketService
	openEventChans  []chan *WSSession
	closeEventChans []chan *WSSession
}

// Create a new WebsocketManager.
func NewSessionManager(maxSessions uint, wss WebSocketService) *SessionManager {
	return &SessionManager{
		maxSessions:     maxSessions,
		activeSessions:  make(map[uint]*WSSession),
		idPool:          NewIdPool(maxSessions),
		mtx:             &sync.Mutex{},
		service:         wss,
		openEventChans:  []chan *WSSession{},
		closeEventChans: []chan *WSSession{},
	}
}

// TODO 
func (this *SessionManager) Shutdown() {
	this.activeSessions = nil
}

// Add a listener to session open events.
func (this *SessionManager) SessionOpenEventChannel() <-chan *WSSession {
	lChan := make(chan *WSSession, 1)
	this.openEventChans = append(this.openEventChans, lChan)
	return lChan
}

// Remove a listener from session open events.
func (this *SessionManager) RemoveSessionOpenEventChannel(lChan chan *WSSession) bool {
	ec := this.openEventChans
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
func (this *SessionManager) SessionCloseEventChannel() <-chan *WSSession {
	lChan := make(chan *WSSession, 1)
	this.closeEventChans = append(this.closeEventChans, lChan)
	return lChan
}

// Remove a listener from session close events.
func (this *SessionManager) RemoveSessionCloseEventChannel(lChan chan *WSSession) bool {
	ec := this.closeEventChans
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
func (this *SessionManager) notifyOpened(session *WSSession) {
	for _, lChan := range this.openEventChans {
		lChan <- session
	}
}

// Used to notify all observers that a new session was closed.
func (this *SessionManager) notifyClosed(session *WSSession) {
	for _, lChan := range this.closeEventChans {
		lChan <- session
	}
}

// Creates a new session and adds it to the manager.
func (this *SessionManager) createSession(wsConn *websocket.Conn) (*WSSession, error) {
	// Check that the capacity hasn't been exceeded.
	this.mtx.Lock()
	defer this.mtx.Unlock()
	if this.atCapacity() {
		return nil, fmt.Errorf("Already at capacity")
	}

	// Create and start
	newId, _ := this.idPool.GetId()
	conn := &WSSession{
		sessionManager: this,
		id:             newId,
		wsConn:         wsConn,
		writeChan:      make(chan []byte, writeChanBufferSize),
		writeCloseChan: make(chan struct{}),
		service:        this.service,
	}
	this.activeSessions[conn.id] = conn
	return conn, nil
}

// Remove a session from the list.
func (this *SessionManager) removeSession(id uint) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	// Check that it exists.
	_, ok := this.activeSessions[id]
	if ok {
		delete(this.activeSessions, id)
		this.idPool.ReleaseId(id)
	}
}

// True if the number of active connections is at the maximum.
func (this *SessionManager) atCapacity() bool {
	return len(this.activeSessions) >= int(this.maxSessions)
}
