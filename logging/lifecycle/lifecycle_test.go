package lifecycle

import (
	"os"
	"testing"

	"bufio"

	. "github.com/hyperledger/burrow/logging/config"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/log15"
)

func TestNewLoggerFromLoggingConfig(t *testing.T) {
	reader := CaptureStderr(t, func() {
		logger, err := NewLoggerFromLoggingConfig(nil)
		assert.NoError(t, err)
		logger.Info("Quick", "Test")
	})
	line, _, err := reader.ReadLine()
	assert.NoError(t, err)
	lineString := string(line)
	assert.NotEmpty(t, lineString)
}

func TestCaptureTendermintLog15Output(t *testing.T) {
	reader := CaptureStderr(t, func() {
		loggingConfig := &LoggingConfig{
			RootSink: Sink().
				SetOutput(StderrOutput().SetFormat("logfmt")).
				SetTransform(FilterTransform(ExcludeWhenAllMatch,
					"log_channel", "Trace",
				)),
		}
		outputLogger, err := NewLoggerFromLoggingConfig(loggingConfig)
		assert.NoError(t, err)
		CaptureTendermintLog15Output(outputLogger)
		log15Logger := log15.New()
		log15Logger.Info("bar", "number_of_forks", 2)
	})
	line, _, err := reader.ReadLine()
	assert.NoError(t, err)
	assert.Contains(t, string(line), "number_of_forks=2")
	assert.Contains(t, string(line), "message=bar")
}

func CaptureStderr(t *testing.T, runner func()) *bufio.Reader {
	stderr := os.Stderr
	defer func() {
		os.Stderr = stderr
	}()
	r, w, err := os.Pipe()
	assert.NoError(t, err, "Couldn't make fifo")
	os.Stderr = w

	runner()

	return bufio.NewReader(r)
}
