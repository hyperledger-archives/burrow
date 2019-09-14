package native

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultNatives(t *testing.T) {
	ns, err := DefaultNatives()
	require.NoError(t, err)
	fmt.Println(ns)
}
