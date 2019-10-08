// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package loggers

import (
	"testing"

	"github.com/hyperledger/burrow/logging/structure"
	"github.com/stretchr/testify/assert"
)

func TestVectorValuedLogger(t *testing.T) {
	logger := newTestLogger()
	vvl := VectorValuedLogger(logger)
	vvl.Log("foo", "bar", "seen", 1, "seen", 3, "seen", 2)
	lls, err := logger.logLines(1)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"foo", "bar", "seen", structure.Vector{1, 3, 2}},
		lls[0])
}
