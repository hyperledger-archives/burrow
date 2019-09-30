package native

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultNatives(t *testing.T) {
	_, err := DefaultNatives()
	require.NoError(t, err)
}
