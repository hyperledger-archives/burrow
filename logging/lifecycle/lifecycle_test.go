package lifecycle

import (
	"runtime"
	"testing"
	"time"
)

func TestNewStdErrLogger(t *testing.T) {
	logger := NewStdErrLogger()
	logger.Info("Quick", "Test")
	time.Sleep(time.Second)
	runtime.Gosched()
}
