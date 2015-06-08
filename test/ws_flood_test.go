package test

import (
	"fmt"
	"github.com/androlo/blockchain_rpc/server"
	"github.com/androlo/blockchain_rpc/test/client"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"time"
)

const CONNS = 100
const MESSAGES = 10

// To keep track of new websocket sessions on the server.
type SessionCounter struct {
	opened int
	closed int
}

func (this *SessionCounter) Run(oChan, cChan <-chan *server.WSSession) {
	go func() {
		for {
			select {
			case <-oChan:
				fmt.Println("Opened")
				this.opened++
				break
			case <-cChan:
				fmt.Println("Closed")
				this.closed++
				break
			}
		}
	}()
}

func (this *SessionCounter) Report() (int, int, int) {
	return this.opened, this.closed, this.opened - this.closed
}

// Coarse flood testing just to ensure that websocket server
// does not crash.
func TestWsFlooding(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// New websocket server.
	wsServer := NewScumsocketServer(CONNS)
	
	// Keep track of sessions.
	sc := &SessionCounter{}
	
	// Register the observer.
	oChan := wsServer.SessionManager().SessionOpenEventChannel()
	cChan := wsServer.SessionManager().SessionCloseEventChannel()

	sc.Run(oChan, cChan)

	serveProcess := NewServeScumSocket(wsServer)
	errServe := serveProcess.Start()
	assert.NoError(t, errServe, "ScumSocketed!")

	// Run
	errRun := runWs()

	errStop := serveProcess.Stop(time.Millisecond * 100)
	assert.NoError(t, errRun, "ScumSocketed!")
	assert.NoError(t, errStop, "ScumSocketed!")
	o, c, a := sc.Report() 
	assert.Equal(t, o, CONNS, "Server registered '%d' opened conns out of '%d'", o, CONNS)
	assert.Equal(t, c, CONNS, "Server registered '%d' closed conns out of '%d'", c, CONNS)
	assert.Equal(t, a, 0, "Server registered '%d' conns still active after closing all.", a)

	fmt.Printf("WebSocket test: A total of %d messages sent succesfully over %d parallel websocket connections.\n", CONNS*MESSAGES, CONNS)
}

func runWs() error {
	doneChan := make(chan bool)
	errChan := make(chan error)
	for i := 0; i < CONNS; i++ {
		go wsClient(doneChan, errChan)
	}
	runners := 0
	for runners < CONNS {
		select {
		case _ = <-doneChan:
			runners++
		case err := <-errChan:
			return err
		}
	}
	return nil
}

func wsClient(doneChan chan bool, errChan chan error) {
	client := client.NewWSClient("ws://localhost:31337/scumsocket")
	_, err := client.Dial()
	if err != nil {
		errChan <- err
		return
	}
	readChan := client.Read()
	i := 0
	start := time.Now()
	for i < MESSAGES {
		client.WriteMsg([]byte("test"))
		<-readChan
		i++
	}
	dur := (time.Since(start).Nanoseconds()) / 1e6
	fmt.Printf("Time taken for %d round trips: %d ms\n", MESSAGES, dur)

	client.Close()
	time.Sleep(time.Millisecond * 500)
	doneChan <- true
}
