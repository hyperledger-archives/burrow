package abi

import (
	"testing"

	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventSpec_Get(t *testing.T) {
	spec, err := ReadSpec(solidity.Abi_EventEmitter)
	require.NoError(t, err)
	eventSpec := spec.EventsByName["ManyTypes2"]

	v, ok := eventSpec.Get("Name")
	require.True(t, ok)
	assert.Equal(t, "ManyTypes2", v)

	v, ok = eventSpec.Get("Inputs")
	require.True(t, ok)
	assert.Equal(t, eventSpec.Inputs, v)
}
