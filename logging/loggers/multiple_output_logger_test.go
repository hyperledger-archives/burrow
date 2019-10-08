// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package loggers

import (
	"testing"

	"github.com/hyperledger/burrow/logging/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewMultipleOutputLogger(t *testing.T) {
	a := newErrorLogger("error a")
	b := newErrorLogger("error b")
	mol := NewMultipleOutputLogger(a, b)
	logLine := []interface{}{"msg", "hello"}
	errLog := mol.Log(logLine...)
	expected := [][]interface{}{logLine}
	logLineA, err := a.logLines(1)
	assert.NoError(t, err)
	logLineB, err := b.logLines(1)
	assert.NoError(t, err)
	assert.Equal(t, expected, logLineA)
	assert.Equal(t, expected, logLineB)
	assert.IsType(t, errors.MultipleErrors{}, errLog)
}
