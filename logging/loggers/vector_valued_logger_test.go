package loggers

import (
	"testing"

	. "github.com/eris-ltd/eris-db/util/slice"
	"github.com/stretchr/testify/assert"
)

func TestVectorValuedLogger(t *testing.T) {
	logger := newTestLogger()
	vvl := VectorValuedLogger(logger)
	vvl.Log("foo", "bar", "seen", 1, "seen", 3, "seen", 2)

	assert.Equal(t, Slice("foo", "bar", "seen", Slice(1, 3, 2)),
		logger.logLines[0])
}
