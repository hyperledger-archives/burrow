package loggers

import (
	"github.com/eapache/channels"
	kitlog "github.com/go-kit/kit/log"
)

const (
	LoggingRingBufferCap channels.BufferCap = 100
)

type ChannelLogger struct {
	ch channels.Channel
}

var _ kitlog.Logger = (*ChannelLogger)(nil)

// Creates a Logger that uses a uses a non-blocking channel.
//
// We would like calls to Log to never block so we use a channel implementation
// that is non-blocking on writes and is able to be so by using a finite ring
// buffer.
func newChannelLogger() *ChannelLogger {
	return &ChannelLogger{
		ch: channels.NewRingChannel(LoggingRingBufferCap),
	}
}

func (cl *ChannelLogger) Log(keyvals ...interface{}) error {
	cl.ch.In() <- keyvals
	// We don't have a way to pass on any logging errors, but that's okay: Log is
	// a maximal interface and the error return type is only there for special
	// cases.
	return nil
}

// Read a log line by waiting until one is available and returning it
func (cl *ChannelLogger) WaitReadLogLine() []interface{} {
	log := <-cl.ch.Out()
	// We are passing slices of interfaces down this channel (go-kit log's Log
	// interface type), a panic is the right thing to do if this type assertion
	// fails.
	return log.([]interface{})
}

// Tries to read a log line from the channel buffer or returns nil if none is
// immediately available
func (cl *ChannelLogger) ReadLogLine() []interface{} {
	select {
	case log := <-cl.ch.Out():
		// See WaitReadLogLine
		return log.([]interface{})
	default:
		return nil
	}
}

// Enters an infinite loop that will drain any log lines from the passed logger.
//
// Exits if the channel is closed.
func (cl *ChannelLogger) DrainChannelToLogger(logger kitlog.Logger) {
	for cl.ch.Out() != nil {
		logger.Log(cl.WaitReadLogLine()...)
	}
}

// Wraps an underlying Logger baseLogger to provide a Logger that is
// is non-blocking on calls to Log.
func NonBlockingLogger(logger kitlog.Logger) *ChannelLogger {
	cl := newChannelLogger()
	go cl.DrainChannelToLogger(logger)
	return cl
}
