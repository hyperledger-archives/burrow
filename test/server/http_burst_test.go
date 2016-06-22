package server

import (
	// "fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

const (
	HTTP_MESSAGES = 300
)

// Send a burst of GET messages to the server.
func TestHttpFlooding(t *testing.T) {
	serveProcess, err := NewServeScumbag()
	assert.NoError(t, err, "Error creating new Server")
	errSS := serveProcess.Start()
	assert.NoError(t, errSS, "Scumbag-ed!")
	t.Logf("Flooding http requests.")
	for i := 0; i < 3; i++ {
		err := runHttp()
		assert.NoError(t, err)
		time.Sleep(200 * time.Millisecond)
	}
	stopC := serveProcess.StopEventChannel()
	errStop := serveProcess.Stop(0)
	<-stopC
	assert.NoError(t, errStop, "Scumbag-ed!")
}

func runHttp() error {
	c := 0
	for c < HTTP_MESSAGES {
		resp, errG := http.Get("http://localhost:31400/scumbag")
		if errG != nil {
			return errG
		}
		c++
		resp.Body.Close()
	}
	return nil
}
