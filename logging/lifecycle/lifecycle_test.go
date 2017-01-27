package lifecycle

import (
	"testing"
	"time"
	"runtime"
)

func TestNewStdErrLogger(t *testing.T) {
	logger := NewStdErrLogger()
	logger.Info("Quick", "Test")
	time.Sleep(time.Second)
	runtime.Gosched()
}
