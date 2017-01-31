package loggers

import (
	"runtime"
	"testing"
	"time"

	"github.com/eris-ltd/eris-db/logging/structure"
	kitlog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestMultipleChannelLogger(t *testing.T) {
	boringLogger, interestingLogger := newTestLogger(), newTestLogger()
	mcl := kitlog.NewContext(MultipleChannelLogger(map[string]kitlog.Logger{
		"Boring":      boringLogger,
		"Interesting": interestingLogger,
	}))
	err := mcl.With("time", kitlog.Valuer(func() interface{} { return "aa" })).
		Log(structure.ChannelKey, "Boring", "foo", "bar")
	assert.NoError(t, err, "Should log without an error")
	// Wait for channel to drain
	time.Sleep(time.Second)
	runtime.Gosched()
	assert.Equal(t, []interface{}{"time", "aa", structure.ChannelKey, "Boring",
		"foo", "bar"},
		boringLogger.logLines[0])
}
