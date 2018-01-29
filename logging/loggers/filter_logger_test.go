package loggers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterLogger(t *testing.T) {
	testLogger := NewChannelLogger(100)
	filterLogger := FilterLogger(testLogger, func(keyvals []interface{}) bool {
		return len(keyvals) > 0 && keyvals[0] == "Spoon"
	})
	filterLogger.Log("Fish", "Present")
	filterLogger.Log("Spoon", "Present")
	assert.Equal(t, [][]interface{}{{"Fish", "Present"}}, testLogger.FlushLogLines())
}
