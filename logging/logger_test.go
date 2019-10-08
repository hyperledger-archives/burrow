// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	bufLogger := log.NewLogfmtLogger(&buf)
	logger := NewLogger(bufLogger)
	logger = logger.With("foo", "bar")
	logger.Trace.Log("hello", "barry")
	require.Equal(t, "log_channel=Trace foo=bar hello=barry\n", buf.String())
}

func TestNewNoopInfoTraceLogger(t *testing.T) {
	logger := NewNoopLogger()
	logger.Trace.Log("goodbye", "trevor")
}
