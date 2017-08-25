package lifecycle

import (
	"os"
	"testing"

	"bufio"

	"github.com/stretchr/testify/assert"
)

func TestNewLoggerFromLoggingConfig(t *testing.T) {
	stderr := os.Stderr
	defer func() {
		os.Stderr = stderr
	}()
	r, w, err := os.Pipe()
	assert.NoError(t, err, "Couldn't make fifo")
	os.Stderr = w
	logger, err := NewLoggerFromLoggingConfig(nil)
	assert.NoError(t, err)
	logger.Info("Quick", "Test")
	reader := bufio.NewReader(r)
	assert.NoError(t, err)
	line, _, err := reader.ReadLine()
	assert.NoError(t, err)
	// This test shouldn't really depend on colour codes, if you find yourself
	// changing it then assert.NotEmpty() should do
	assert.Contains(t, string(line), "\x1b[34mQuick\x1b[0m=Test")
}
