package loggers

import (
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestChannelLogger(t *testing.T) {
	cl := newChannelLogger()

	// Push a larger number of log messages than will fit into ring buffer
	for i := 0; i < int(LoggingRingBufferCap)+10; i++ {
		cl.Log("log line", i)
	}

	// Observe that oldest 10 messages are overwritten (so first message is 10)
	for i := 0; i < int(LoggingRingBufferCap); i++ {
		ll := cl.WaitReadLogLine()
		assert.Equal(t, 10+i, ll[1])
	}

	assert.Nil(t, cl.ReadLogLine(), "Since we have drained the buffer there "+
		"should be no more log lines.")
}

func TestBlether(t *testing.T) {
	var bs []byte
	ext := append(bs)
	fmt.Println(ext)
}
