package loggers

import "errors"

type testLogger struct {
	logLines [][]interface{}
	err      error
}

func newErrorLogger(errMessage string) *testLogger {
	return &testLogger{err: errors.New(errMessage)}
}

func newTestLogger() *testLogger {
	return &testLogger{}
}

func (tl *testLogger) Log(keyvals ...interface{}) error {
	tl.logLines = append(tl.logLines, keyvals)
	return tl.err
}
