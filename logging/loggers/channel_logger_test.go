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

package loggers

import (
	"testing"

	"time"

	"github.com/eapache/channels"
	"github.com/stretchr/testify/assert"
)

func TestChannelLogger(t *testing.T) {
	loggingRingBufferCap := channels.BufferCap(5)
	cl := NewChannelLogger(loggingRingBufferCap)

	// Push a larger number of log messages than will fit into ring buffer
	for i := 0; i < int(loggingRingBufferCap)+10; i++ {
		cl.Log("log line", i)
	}

	// Observe that oldest 10 messages are overwritten (so first message is 10)
	for i := 0; i < int(loggingRingBufferCap); i++ {
		ll := cl.WaitReadLogLine()
		assert.Equal(t, 10+i, ll[1])
	}

	assert.Nil(t, cl.ReadLogLine(), "Since we have drained the buffer there "+
		"should be no more log lines.")
}

func TestChannelLogger_Reset(t *testing.T) {
	loggingRingBufferCap := channels.BufferCap(5)
	cl := NewChannelLogger(loggingRingBufferCap)
	for i := 0; i < int(loggingRingBufferCap); i++ {
		cl.Log("log line", i)
	}
	cl.Reset()
	for i := 0; i < int(loggingRingBufferCap); i++ {
		cl.Log("log line", i)
	}
	for i := 0; i < int(loggingRingBufferCap); i++ {
		ll := cl.WaitReadLogLine()
		assert.Equal(t, i, ll[1])
	}
	assert.Nil(t, cl.ReadLogLine(), "Since we have drained the buffer there "+
		"should be no more log lines.")
}

func TestNonBlockingLogger(t *testing.T) {
	tl := newTestLogger()
	nbl := NonBlockingLogger(tl)
	nbl.Log("Foo", "Bar")
	nbl.Log("Baz", "Bur")
	nbl.Log("Badger", "Romeo")
	time.Sleep(time.Second)

	lls, err := tl.logLines(3)
	assert.NoError(t, err)
	assert.Equal(t, logLines("Foo", "Bar", "",
		"Baz", "Bur", "",
		"Badger", "Romeo"), lls)
}
