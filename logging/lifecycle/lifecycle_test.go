package lifecycle

import (
	"os"
	"testing"

	"bufio"

	"github.com/stretchr/testify/assert"
)

func TestNewLoggerFromLoggingConfig(t *testing.T) {
	reader := CaptureStderr(t, func() {
		logger, err := NewLoggerFromLoggingConfig(nil)
		assert.NoError(t, err)
		logger.Info.Log("Quick", "Test")
	})
	line, _, err := reader.ReadLine()
	assert.NoError(t, err)
	lineString := string(line)
	assert.NotEmpty(t, lineString)
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
