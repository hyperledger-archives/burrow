package loggers

import (
	"github.com/eapache/channels"
	kitlog "github.com/go-kit/kit/log"
	"sync"
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
// Because it holds a refereence to its output it can also be used to coordinate
// Flushing of the buffer to the output logger only in exceptional circumstances
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
	if cl.passthrough {
		err = cl.outputLogger.Log(keyvals...)
	}
	return err
}

func (cl *CaptureLogger) SetPassthrough(passthrough bool) {
	cl.RWMutex.Lock()
	cl.passthrough = passthrough
	cl.RWMutex.Unlock()
}

func (cl *CaptureLogger) Passthrough() bool {
	cl.RWMutex.RLock()
	passthrough := cl.passthrough
	cl.RWMutex.RUnlock()
	return passthrough
}

// Flushes every log line available in the buffer at the time of calling
// to the OutputLogger and returns. Does not block indefinitely.
func (cl *CaptureLogger) Flush() {
	cl.bufferLogger.Flush(cl.outputLogger)
}

func (cl *CaptureLogger) OutputLogger() kitlog.Logger {
	return cl.outputLogger
}

func (cl *CaptureLogger) BufferLogger() *ChannelLogger {
	return cl.bufferLogger
}
