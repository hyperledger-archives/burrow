// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package loggers

import (
	"testing"

	"time"

	"fmt"

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
	nbl, _ := NonBlockingLogger(tl)
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

func TestNonBlockingLoggerErrors(t *testing.T) {
	el := newErrorLogger("Should surface")
	nbl, errCh := NonBlockingLogger(el)
	nbl.Log("failure", "true")
	assert.Equal(t, "Should surface",
		fmt.Sprintf("%s", <-errCh.Out()))
}
