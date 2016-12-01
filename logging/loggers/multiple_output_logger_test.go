package loggers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMultipleOutputLogger(t *testing.T) {
	a, b := newErrorLogger("error a"), newErrorLogger("error b")
	mol := NewMultipleOutputLogger(a, b)
	logLine := []interface{}{"msg", "hello"}
	err := mol.Log(logLine...)
	expected := [][]interface{}{logLine}
	assert.Equal(t, expected, a.logLines)
	assert.Equal(t, expected, b.logLines)
	assert.IsType(t, multipleErrors{}, err)
}
