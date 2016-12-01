package loggers

import (
	"runtime"
	"testing"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestMultipleChannelLogger(t *testing.T) {
	boringLogger, interestingLogger := newTestLogger(), newTestLogger()
	mcl := NewMultipleChannelLogger(map[string]kitlog.Logger{
		"Boring":      boringLogger,
		"Interesting": interestingLogger,
	})
	err := mcl.With("time", kitlog.Valuer(func() interface{} { return "aa" })).
		Log("Boring", "foo", "bar")
	assert.NoError(t, err, "Should log without an error")
	// Wait for channel to drain
	time.Sleep(time.Second)
	runtime.Gosched()
	assert.Equal(t, []interface{}{"time", "aa", "foo", "bar"},
		boringLogger.logLines[0])
}
