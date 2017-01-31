package loggers

import "errors"

type testLogger struct {
	cl  *ChannelLogger
	err error
}

func (el *testLogger) empty() bool {
	return el.cl.BufferLength() == 0
}

func (el *testLogger) logLines() [][]interface{} {
	return el.cl.WaitLogLines()
}

func (el *testLogger) Log(keyvals ...interface{}) error {
	el.cl.Log(keyvals...)
	return el.err
}

func newErrorLogger(errMessage string) *testLogger {
	return &testLogger{
		cl:  NewChannelLogger(100),
		err: errors.New(errMessage),
	}
}

func newTestLogger() *testLogger {
	return &testLogger{
		cl:  NewChannelLogger(100),
		err: nil,
	}
}
