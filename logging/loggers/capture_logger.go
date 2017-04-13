package loggers

import (
	"sync"

	"github.com/eapache/channels"
	kitlog "github.com/go-kit/kit/log"
)

type CaptureLogger struct {
	bufferLogger *ChannelLogger
	outputLogger kitlog.Logger
	passthrough  bool
	sync.RWMutex
}

var _ kitlog.Logger = (*CaptureLogger)(nil)

// Capture logger captures output set to it into a buffer logger and retains
// a reference to an output logger (the logger whose input it is capturing).
// It can optionally passthrough logs to the output logger.
// Because it holds a reference to its output it can also be used to coordinate
// Flushing of the buffer to the output logger in exceptional circumstances only
func NewCaptureLogger(outputLogger kitlog.Logger, bufferCap channels.BufferCap,
	passthrough bool) *CaptureLogger {
	return &CaptureLogger{
		bufferLogger: NewChannelLogger(bufferCap),
		outputLogger: outputLogger,
		passthrough:  passthrough,
	}
}

func (cl *CaptureLogger) Log(keyvals ...interface{}) error {
	err := cl.bufferLogger.Log(keyvals...)
	if cl.Passthrough() {
		err = cl.outputLogger.Log(keyvals...)
	}
	return err
}

// Sets whether the CaptureLogger is forwarding log lines sent to it through
// to its output logger. Concurrently safe.
func (cl *CaptureLogger) SetPassthrough(passthrough bool) {
	cl.RWMutex.Lock()
	cl.passthrough = passthrough
	cl.RWMutex.Unlock()
}

// Gets whether the CaptureLogger is forwarding log lines sent to through to its
// OutputLogger. Concurrently Safe.
func (cl *CaptureLogger) Passthrough() bool {
	cl.RWMutex.RLock()
	passthrough := cl.passthrough
	cl.RWMutex.RUnlock()
	return passthrough
}

// Flushes every log line available in the buffer at the time of calling
// to the OutputLogger and returns. Does not block indefinitely.
//
// Note: will remove log lines from buffer so they will not be produced on any
// subsequent flush of buffer
func (cl *CaptureLogger) Flush() {
	cl.bufferLogger.Flush(cl.outputLogger)
}

// Flushes every log line available in the buffer at the time of calling
// to a slice and returns it. Does not block indefinitely.
//
// Note: will remove log lines from buffer so they will not be produced on any
// subsequent flush of buffer
func (cl *CaptureLogger) FlushLogLines() [][]interface{} {
	return cl.bufferLogger.FlushLogLines()
}

// The OutputLogger whose input this CaptureLogger is capturing
func (cl *CaptureLogger) OutputLogger() kitlog.Logger {
	return cl.outputLogger
}

// The BufferLogger where the input into these CaptureLogger is stored in a ring
// buffer of log lines.
func (cl *CaptureLogger) BufferLogger() *ChannelLogger {
	return cl.bufferLogger
}
