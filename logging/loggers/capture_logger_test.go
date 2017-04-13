package loggers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlushCaptureLogger(t *testing.T) {
	outputLogger := newTestLogger()
	cl := NewCaptureLogger(outputLogger, 100, false)
	buffered := 50
	for i := 0; i < buffered; i++ {
		cl.Log("Foo", "Bar", "Index", i)
	}
	assert.True(t, outputLogger.empty())

	// Flush the ones we bufferred
	cl.Flush()
	_, err := outputLogger.logLines(buffered)
	assert.NoError(t, err)
}

func TestTeeCaptureLogger(t *testing.T) {
	outputLogger := newTestLogger()
	cl := NewCaptureLogger(outputLogger, 100, true)
	buffered := 50
	for i := 0; i < buffered; i++ {
		cl.Log("Foo", "Bar", "Index", i)
	}
	// Check passthrough to output
	ll, err := outputLogger.logLines(buffered)
	assert.NoError(t, err)
	assert.Equal(t, ll, cl.BufferLogger().FlushLogLines())

	cl.SetPassthrough(false)
	buffered = 110
	for i := 0; i < buffered; i++ {
		cl.Log("Foo", "Bar", "Index", i)
	}
	assert.True(t, outputLogger.empty())

	cl.Flush()
	_, err = outputLogger.logLines(100)
	assert.NoError(t, err)
	_, err = outputLogger.logLines(1)
	// Expect timeout
	assert.Error(t, err)
}
