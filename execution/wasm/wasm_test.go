package wasm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/hyperledger/burrow/execution/native"

	"github.com/hyperledger/burrow/execution/exec"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/evm/abi"

	"github.com/hyperledger/burrow/crypto"
	"github.com/stretchr/testify/require"
)

func TestStaticCallWithValue(t *testing.T) {
	cache := acmstate.NewMemoryState()

	params := engine.CallParams{
		Origin: crypto.ZeroAddress,
		Caller: crypto.ZeroAddress,
		Callee: crypto.ZeroAddress,
		Input:  []byte{},
		Value:  *big.NewInt(0),
		Gas:    big.NewInt(1000),
	}

	vm := Default()
	blockchain := new(engine.TestBlockchain)
	eventSink := exec.NewNoopEventSink()

	// run constructor
	runtime, cerr := vm.Execute(cache, blockchain, eventSink, params, Bytecode_storage_test)
	require.NoError(t, cerr)

	// run getFooPlus2
	spec, err := abi.ReadSpec(Abi_storage_test)
	require.NoError(t, err)
	calldata, _, err := spec.Pack("getFooPlus2")

	params.Input = calldata

	returndata, cerr := vm.Execute(cache, blockchain, eventSink, params, runtime)
	require.NoError(t, cerr)

	data := abi.GetPackingTypes(spec.Functions["getFooPlus2"].Outputs)

	err = spec.Unpack(returndata, "getFooPlus2", data...)
	require.NoError(t, err)
	returnValue := *data[0].(*uint64)
	var expected uint64
	expected = 104
	require.Equal(t, expected, returnValue)

	// call incFoo
	calldata, _, err = spec.Pack("incFoo")

	params.Input = calldata

	returndata, cerr = vm.Execute(cache, blockchain, eventSink, params, runtime)
	require.NoError(t, cerr)

	require.Equal(t, returndata, []byte{})

	// run getFooPlus2
	calldata, _, err = spec.Pack("getFooPlus2")
	require.NoError(t, err)

	params.Input = calldata

	returndata, cerr = vm.Execute(cache, blockchain, eventSink, params, runtime)
	require.NoError(t, cerr)

	spec.Unpack(returndata, "getFooPlus2", data...)
	expected = 105
	returnValue = *data[0].(*uint64)

	require.Equal(t, expected, returnValue)
}

func TestCREATE(t *testing.T) {
	cache := acmstate.NewMemoryState()

	params := engine.CallParams{
		Origin: crypto.ZeroAddress,
		Caller: crypto.ZeroAddress,
		Callee: crypto.ZeroAddress,
		Input:  []byte{},
		Value:  *big.NewInt(0),
		Gas:    big.NewInt(1000),
	}

	vm := New(engine.Options{Natives: native.MustDefaultNatives()})
	blockchain := new(engine.TestBlockchain)
	eventSink := exec.NewNoopEventSink()

	// run constructor
	runtime, cerr := vm.Execute(cache, blockchain, eventSink, params, CREATETest)
	require.NoError(t, cerr)

	// run createChild
	spec, err := abi.ReadSpec(Abi_CREATETest)
	require.NoError(t, err)
	calldata, _, err := spec.Pack("createChild")

	params.Input = calldata

	_, rerr := vm.Execute(cache, blockchain, eventSink, params, runtime)
	vm.options.Nonce = []byte{0xff}
	_, rerr = vm.Execute(cache, blockchain, eventSink, params, runtime)
	require.NoError(t, rerr)

	// get created child
	calldata, _, err = spec.Pack("getChild", "0")
	require.NoError(t, err)

	params.Input = calldata

	res, rerr := vm.Execute(cache, blockchain, eventSink, params, runtime)
	require.NoError(t, rerr)
	require.Equal(t, "000000000000000000000000ef2fb521372225b89169ba60500142f68ebd82d3", hex.EncodeToString(res))

	calldata, _, err = spec.Pack("getChild", "1")
	require.NoError(t, err)

	params.Input = calldata

	res, rerr = vm.Execute(cache, blockchain, eventSink, params, runtime)
	require.NoError(t, rerr)
	require.Equal(t, "00000000000000000000000089686394a7cf94be0aa48ae593fe3cad5cbdbace", hex.EncodeToString(res))
}

func TestSelfDestruct(t *testing.T) {
	cache := acmstate.NewMemoryState()

	params := engine.CallParams{
		Origin: crypto.ZeroAddress,
		Caller: crypto.ZeroAddress,
		Callee: crypto.ZeroAddress,
		Input:  []byte{},
		Value:  *big.NewInt(0),
		Gas:    big.NewInt(1000),
	}

	vm := New(engine.Options{Natives: native.MustDefaultNatives()})
	blockchain := new(engine.TestBlockchain)
	eventSink := exec.NewNoopEventSink()

	// run constructor
	runtime, cerr := vm.Execute(cache, blockchain, eventSink, params, CREATETest)
	require.NoError(t, cerr)
	require.Equal(t, 1, len(cache.Accounts))

	// run selfdestruct
	spec, err := abi.ReadSpec(Abi_CREATETest)
	require.NoError(t, err)
	calldata, _, err := spec.Pack("close")

	params.Input = calldata

	_, rerr := vm.Execute(cache, blockchain, eventSink, params, runtime)
	require.NoError(t, rerr)

	require.Equal(t, 0, len(cache.Accounts))
}

func TestMisc(t *testing.T) {
	cache := acmstate.NewMemoryState()

	params := engine.CallParams{
		Origin: crypto.ZeroAddress,
		Caller: crypto.ZeroAddress,
		Callee: crypto.ZeroAddress,
		Input:  []byte{},
		Value:  *big.NewInt(0),
		Gas:    big.NewInt(1000),
	}

	vm := New(engine.Options{Natives: native.MustDefaultNatives()})
	blockchain := new(engine.TestBlockchain)
	eventSink := exec.NewNoopEventSink()

	// run constructor
	runtime, cerr := vm.Execute(cache, blockchain, eventSink, params, CREATETest)
	require.NoError(t, cerr)

	// run txGasPrice
	spec, err := abi.ReadSpec(Abi_CREATETest)
	require.NoError(t, err)
	calldata, _, err := spec.Pack("txPrice")

	params.Input = calldata

	res, rerr := vm.Execute(cache, blockchain, eventSink, params, runtime)
	require.NoError(t, rerr)
	require.Equal(t, "0000000000000000000000000000000000000000000000000000000000000001", hex.EncodeToString(res))

	// run blockDifficulty
	spec, err = abi.ReadSpec(Abi_CREATETest)
	require.NoError(t, err)
	calldata, _, err = spec.Pack("blockDifficulty")

	params.Input = calldata

	res, rerr = vm.Execute(cache, blockchain, eventSink, params, runtime)
	require.NoError(t, rerr)
	require.Equal(t, "0000000000000000000000000000000000000000000000000000000000000001", hex.EncodeToString(res))
}

func blockHashGetter(height uint64) []byte {
	return binary.LeftPadWord256([]byte(fmt.Sprintf("block_hash_%d", height))).Bytes()
}
