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
	// "fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
