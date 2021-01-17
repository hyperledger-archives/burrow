// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"testing"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"

	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	"github.com/stretchr/testify/assert"
)

func TestSNativeContractDescription1a(t *testing.T) {
	signatory := []byte("4fbdd7d412e5dad21d0c4add8b3cdc70ab65a5b3")
	signature := []byte("e8cd78a90152a1549cfea3dcc8556e25618c67085c39e0b95e3d1ec6dbe9339525e7c61e005954fa67ce0c102cee434d1b111c4b0c52bad0fdc2663ca09451ac1b")
	messageDigest := []byte("21400fb863b0ee7392af02491b6ca5fb87f68ad6af4867b104282dd8fe8b52e4")

	st := acmstate.NewMemoryState()
	caller := &acm.Account{
		Address: crypto.Address{1, 1, 1},
	}
	function := Precompiles.GetByAddress(leftPadAddress(1)).(*Function)
	require.NotNil(t, function, "Could not get function: %s")
	require.NoError(t, st.UpdateAccount(caller))

	state := engine.State{
		CallFrame: engine.NewCallFrame(st),
		EventSink: exec.NewNoopEventSink(),
	}

	funcID := function.Abi().FunctionID
	gas := uint64(10000)

	input := bc.MustSplice(funcID[:], messageDigest, signature)
	params := engine.CallParams{
		Caller: caller.Address,
		Input:  input,
		Gas:    &gas,
	}

	returnValue, err := function.Call(state, params)
	require.NoError(t, err)
	assert.Equal(t, signatory, returnValue)
}
