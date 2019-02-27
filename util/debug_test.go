package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDebugDisabled(t *testing.T) {
	// Double check to disable debug (note: util.Debugf statements ought to also be removed from code in most circumstances)
	require.False(t, debug)
}
