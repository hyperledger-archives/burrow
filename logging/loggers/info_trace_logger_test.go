package loggers

import (
	"os"
	"testing"

	kitlog "github.com/go-kit/kit/log"
)

func TestLogger(t *testing.T) {
	stderrLogger := kitlog.NewLogfmtLogger(os.Stderr)
	logger := NewInfoTraceLogger(stderrLogger, stderrLogger)
	logger.Trace("hello", "barry")
}
