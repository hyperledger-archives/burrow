// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"testing"

	"github.com/hyperledger/burrow/execution/evm/abi"

	"github.com/btcsuite/btcd/btcec"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/native"
	"github.com/hyperledger/burrow/execution/solidity"

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
	message := []byte("THIS MESSAGE IS NOT SIGNED")
	digest := crypto.Keccak256(message)
	privateKey, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	sig, err := btcec.SignCompact(btcec.S256(), privateKey, digest, false)
	require.NoError(t, err)

	st := native.NewState(native.MustDefaultNatives(), acmstate.NewMemoryState())
	caller := &acm.Account{
		Address: crypto.Address{1, 1, 1},
	}
	function := native.Precompiles.GetByName("ecrecover").(*native.Function)
	require.NotNil(t, function, "Could not get function: %s")
	require.NoError(t, st.UpdateAccount(caller))

	state := engine.State{
		CallFrame: engine.NewCallFrame(st),
		EventSink: exec.NewNoopEventSink(),
	}

	gas := uint64(10000)

	spec, err := abi.ReadSpec(solidity.Abi_ECRecover)
	require.NoError(t, err)
	funcId := spec.Functions["recoverSigningAddress"].FunctionID
	input := bc.MustSplice(funcId, digest, binary.Int64ToWord256(int64(sig[2*binary.Word256Bytes])), sig[:2*binary.Word256Bytes])

	params := engine.CallParams{
		Caller: caller.Address,
		Input:  input,
		Gas:    &gas,
	}

	vm := New(Options{
		Natives: native.MustDefaultNatives(),
	})
	returnValue, err := vm.Contract(solidity.DeployedBytecode_ECRecover).Call(state, params)
	require.NoError(t, err)
	priv, err := crypto.PrivateKeyFromRawBytes(privateKey.D.Bytes(), crypto.CurveTypeSecp256k1)
	require.NoError(t, err)
	address := priv.GetPublicKey().GetAddress()
	addressOut := crypto.AddressFromWord256(binary.LeftPadWord256(returnValue))
	require.NoError(t, err)
	assert.Equal(t, address, addressOut)
}
