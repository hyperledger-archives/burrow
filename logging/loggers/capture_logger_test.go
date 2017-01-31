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
	ll := outputLogger.logLines()
	assert.Equal(t, buffered, len(ll))
}

func TestTeeCaptureLogger(t *testing.T) {
	outputLogger := newTestLogger()
	cl := NewCaptureLogger(outputLogger, 100, true)
	buffered := 50
	for i := 0; i < buffered; i++ {
		cl.Log("Foo", "Bar", "Index", i)
	}
	// Check passthrough to output
	ll := outputLogger.logLines()
	assert.Equal(t, buffered, len(ll))
	assert.Equal(t, ll, cl.BufferLogger().FlushLogLines())

	cl.SetPassthrough(false)
	buffered = 110
	for i := 0; i < buffered; i++ {
		cl.Log("Foo", "Bar", "Index", i)
	}
	assert.True(t, outputLogger.empty())

	cl.Flush()
	assert.Equal(t, 100, len(outputLogger.logLines()))
}