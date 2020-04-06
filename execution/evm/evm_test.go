// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	. "github.com/hyperledger/burrow/execution/evm/asm"
	. "github.com/hyperledger/burrow/execution/evm/asm/bc"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/native"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmthrgd/go-hex"
)

// Test output is a bit clearer if we /dev/null the logging, but can be re-enabled by uncommenting the below
//var logger, _, _ = lifecycle.NewStdErrLogger()

// Runs a basic loop
func TestEVM(t *testing.T) {
	vm := New(Options{
		Natives: native.MustDefaultNatives(),
	})

	t.Run("BasicLoop", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		blockchain := new(blockchain)
		eventSink := exec.NewNoopEventSink()

		// Create accounts
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "101")

		var gas uint64 = 100000

		bytecode := MustSplice(PUSH1, 0x00, PUSH1, 0x20, MSTORE, JUMPDEST, PUSH2, 0x0F, 0x0F, PUSH1, 0x20, MLOAD,
			SLT, ISZERO, PUSH1, 0x1D, JUMPI, PUSH1, 0x01, PUSH1, 0x20, MLOAD, ADD, PUSH1, 0x20,
			MSTORE, PUSH1, 0x05, JUMP, JUMPDEST)

		start := time.Now()
		output, err := vm.Execute(st, blockchain, eventSink, engine.CallParams{
			Caller: account1,
			Callee: account2,
			Gas:    &gas,
		}, bytecode)
		t.Logf("Output: %v Error: %v\n", output, err)
		t.Logf("Call took: %v", time.Since(start))
		require.NoError(t, err)
	})

	t.Run("SHL", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "101")

		var gas uint64 = 100000

		//Shift left 0
		bytecode := MustSplice(PUSH1, 0x01, PUSH1, 0x00, SHL, return1())
		output, err := call(vm, st, account1, account2, bytecode, nil, &gas)
		value := []byte{0x1}
		expected := LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift left 0
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x00, SHL, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift left 1
		bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0x01, SHL, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x2}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift left 1
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SHL, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x2}
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift left 1
		bytecode = MustSplice(PUSH32, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SHL, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE}

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift left 255
		bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0xFF, SHL, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x80}
		expected = RightPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift left 255
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SHL, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x80}
		expected = RightPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift left 256 (overflow)
		bytecode = MustSplice(PUSH1, 0x01, PUSH2, 0x01, 0x00, SHL, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x00}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift left 256 (overflow)
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH2, 0x01, 0x00, SHL,
			return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x00}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift left 257 (overflow)
		bytecode = MustSplice(PUSH1, 0x01, PUSH2, 0x01, 0x01, SHL, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x00}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("SHR", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "101")

		var gas uint64 = 100000

		//Shift right 0
		bytecode := MustSplice(PUSH1, 0x01, PUSH1, 0x00, SHR, return1())
		output, err := call(vm, st, account1, account2, bytecode, nil, &gas)
		value := []byte{0x1}
		expected := LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift right 0
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x00, SHR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift right 1
		bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0x01, SHR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x00}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift right 1
		bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0x01, SHR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x40}
		expected = RightPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift right 1
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SHR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		expected = []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift right 255
		bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0xFF, SHR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x1}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift right 255
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SHR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x1}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift right 256 (underflow)
		bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH2, 0x01, 0x00, SHR,
			return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x00}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift right 256 (underflow)
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH2, 0x01, 0x00, SHR,
			return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x00}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift right 257 (underflow)
		bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH2, 0x01, 0x01, SHR,
			return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x00}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("SAR", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "101")

		var gas uint64 = 100000

		//Shift arith right 0
		bytecode := MustSplice(PUSH1, 0x01, PUSH1, 0x00, SAR, return1())
		output, err := call(vm, st, account1, account2, bytecode, nil, &gas)
		value := []byte{0x1}
		expected := LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative arith shift right 0
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x00, SAR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift arith right 1
		bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0x01, SAR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x00}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift arith right 1
		bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0x01, SAR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0xc0}
		expected = RightPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift arith right 1
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SAR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift arith right 255
		bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0xFF, SAR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift arith right 255
		bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SAR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift arith right 255
		bytecode = MustSplice(PUSH32, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SAR, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []uint8([]byte{0x00})
		expected = RightPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift arith right 256 (reset)
		bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH2, 0x01, 0x00, SAR,
			return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Alternative shift arith right 256 (reset)
		bytecode = MustSplice(PUSH32, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH2, 0x01, 0x00, SAR,
			return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		value = []byte{0x00}
		expected = LeftPadBytes(value, 32)
		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}

		//Shift arith right 257 (reset)
		bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH2, 0x01, 0x01, SAR,
			return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		expected = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

		assert.Equal(t, expected, output)

		t.Logf("Result: %v == %v\n", output, expected)

		if err != nil {
			t.Fatal(err)
		}
	})

	//Test attempt to jump to bad destination (position 16)
	t.Run("JumpErr", func(t *testing.T) {
		st := acmstate.NewMemoryState()

		// Create accounts
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "2")

		var gas uint64 = 100000

		bytecode := MustSplice(PUSH1, 0x10, JUMP)

		var err error
		ch := make(chan struct{})
		go func() {
			_, err = call(vm, st, account1, account2, bytecode, nil, &gas)
			ch <- struct{}{}
		}()
		tick := time.NewTicker(time.Second * 2)
		select {
		case <-tick.C:
			t.Fatal("VM ended up in an infinite loop from bad jump dest (it took too long!)")
		case <-ch:
			if err == nil {
				t.Fatal("Expected invalid jump dest err")
			}
		}
	})

	// Tests the code for a subcurrency contract compiled by serpent
	t.Run("Subcurrency", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		// Create accounts
		account1 := newAccount(t, st, "1, 2, 3")
		account2 := newAccount(t, st, "3, 2, 1")

		var gas uint64 = 1000

		bytecode := MustSplice(PUSH3, 0x0F, 0x42, 0x40, CALLER, SSTORE, PUSH29, 0x01, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1,
			0x00, CALLDATALOAD, DIV, PUSH4, 0x15, 0xCF, 0x26, 0x84, DUP2, EQ, ISZERO, PUSH2,
			0x00, 0x46, JUMPI, PUSH1, 0x04, CALLDATALOAD, PUSH1, 0x40, MSTORE, PUSH1, 0x40,
			MLOAD, SLOAD, PUSH1, 0x60, MSTORE, PUSH1, 0x20, PUSH1, 0x60, RETURN, JUMPDEST,
			PUSH4, 0x69, 0x32, 0x00, 0xCE, DUP2, EQ, ISZERO, PUSH2, 0x00, 0x87, JUMPI, PUSH1,
			0x04, CALLDATALOAD, PUSH1, 0x80, MSTORE, PUSH1, 0x24, CALLDATALOAD, PUSH1, 0xA0,
			MSTORE, CALLER, SLOAD, PUSH1, 0xC0, MSTORE, CALLER, PUSH1, 0xE0, MSTORE, PUSH1,
			0xA0, MLOAD, PUSH1, 0xC0, MLOAD, SLT, ISZERO, ISZERO, PUSH2, 0x00, 0x86, JUMPI,
			PUSH1, 0xA0, MLOAD, PUSH1, 0xC0, MLOAD, SUB, PUSH1, 0xE0, MLOAD, SSTORE, PUSH1,
			0xA0, MLOAD, PUSH1, 0x80, MLOAD, SLOAD, ADD, PUSH1, 0x80, MLOAD, SSTORE, JUMPDEST,
			JUMPDEST, POP, JUMPDEST, PUSH1, 0x00, PUSH1, 0x00, RETURN)

		data := hex.MustDecodeString("693200CE0000000000000000000000004B4363CDE27C2EB05E66357DB05BC5C88F850C1A0000000000000000000000000000000000000000000000000000000000000005")
		output, err := call(vm, st, account1, account2, bytecode, data, &gas)
		t.Logf("Output: %v Error: %v\n", output, err)
		if err != nil {
			t.Fatal(err)
		}
		require.NoError(t, err)
	})

	//This test case is taken from EIP-140 (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-140.md);
	//it is meant to test the implementation of the REVERT opcode
	t.Run("Revert", func(t *testing.T) {
		st := acmstate.NewMemoryState()

		// Create accounts
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "1, 0, 1")

		key, value := []byte{0x00}, []byte{0x00}
		err := st.SetStorage(account1, LeftPadWord256(key), value)
		require.NoError(t, err)

		var gas uint64 = 100000

		bytecode := MustSplice(PUSH13, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x65, 0x64, 0x20, 0x64, 0x61, 0x74, 0x61,
			PUSH1, 0x00, SSTORE, PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, REVERT)

		/*bytecode := MustSplice(PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61,
		0x67, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, REVERT)*/

		output, err := call(vm, st, account1, account2, bytecode, nil, &gas)
		assert.Error(t, err, "Expected execution reverted error")

		storageVal, err := st.GetStorage(account1, LeftPadWord256(key))
		require.NoError(t, err)
		assert.Equal(t, value, storageVal)

		t.Logf("Output: %v\n", output)
	})

	// Test sending tokens from a contract to another account
	t.Run("SendCall", func(t *testing.T) {
		st := acmstate.NewMemoryState()

		// Create accounts
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "2")
		account3 := newAccount(t, st, "3")

		// account1 will call account2 which will trigger CALL opcode to account3
		addr := account3
		contractCode := callContractCode(addr)

		//----------------------------------------------
		// account2 has insufficient balance, should fail
		txe := runVM(st, account1, account2, contractCode, 100000)
		exCalls := txe.ExceptionalCalls()
		require.Len(t, exCalls, 1)
		require.Equal(t, errors.Codes.InsufficientBalance, errors.GetCode(exCalls[0].Header.Exception))

		//----------------------------------------------
		// give account2 sufficient balance, should pass
		addToBalance(t, st, account2, 100000)
		txe = runVM(st, account1, account2, contractCode, 1000)
		assert.Nil(t, txe.Exception, "Should have sufficient balance")

		//----------------------------------------------
		// insufficient gas, should fail
		txe = runVM(st, account1, account2, contractCode, 100)
		assert.NotNil(t, txe.Exception, "Expected insufficient gas error")
	})

	// Test to ensure that contracts called with STATICCALL cannot modify state
	// as per https://github.com/ethereum/EIPs/blob/master/EIPS/eip-214.md
	t.Run("StaticCallReadOnly", func(t *testing.T) {
		gas1, gas2 := byte(0x1), byte(0x1)
		value := byte(0x69)
		var inOff, inSize, retOff, retSize byte

		logDefault := MustSplice(PUSH1, inSize, PUSH1, inOff)
		testRecipient := native.AddressFromName("1")
		// check all illegal state modifications in child staticcall frame
		for _, illegalContractCode := range []acm.Bytecode{
			MustSplice(PUSH9, "arbitrary", PUSH1, 0x00, SSTORE),
			MustSplice(logDefault, LOG0),
			MustSplice(logDefault, PUSH1, 0x1, LOG1),
			MustSplice(logDefault, PUSH1, 0x1, PUSH1, 0x1, LOG2),
			MustSplice(logDefault, PUSH1, 0x1, PUSH1, 0x1, PUSH1, 0x1, LOG3),
			MustSplice(logDefault, PUSH1, 0x1, PUSH1, 0x1, PUSH1, 0x1, PUSH1, 0x1, LOG4),
			MustSplice(PUSH1, 0x0, PUSH1, 0x0, PUSH1, 0x69, CREATE),
			MustSplice(PUSH20, testRecipient, SELFDESTRUCT),
		} {
			// TODO: CREATE2

			t.Logf("Testing state-modifying bytecode: %v", illegalContractCode.MustTokens())
			st := acmstate.NewMemoryState()
			callee := makeAccountWithCode(t, st, "callee", MustSplice(illegalContractCode, PUSH1, 0x1, return1()))

			// equivalent to CALL, but enforce state immutability for children
			code := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1, inOff,
				PUSH1, value, PUSH20, callee, PUSH2, gas1, gas2, STATICCALL, PUSH1, retSize,
				PUSH1, retOff, RETURN)
			caller := makeAccountWithCode(t, st, "caller", code)

			txe := runVM(st, caller, callee, code, 1000)
			// the topmost caller can never *illegally* modify state
			require.Error(t, txe.Exception)
			require.Equal(t, errors.Codes.IllegalWrite, txe.Exception.ErrorCode(),
				"should get an error from child accounts that st is read only")
		}
	})

	t.Run("StaticCallWithValue", func(t *testing.T) {
		gas1, gas2 := byte(0x1), byte(0x1)
		value := byte(0x69)
		var inOff, inSize, retOff, retSize byte

		st := acmstate.NewMemoryState()

		finalAddress := makeAccountWithCode(t, st, "final", MustSplice(PUSH1, int64(20), return1()))

		// intermediate account CALLs another contract *with* a value
		callee := makeAccountWithCode(t, st, "callee", MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1,
			inOff, PUSH1, value, PUSH20, finalAddress, PUSH2, gas1, gas2, CALL, returnWord()))

		callerCode := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1,
			inOff, PUSH1, value, PUSH20, callee, PUSH2, gas1, gas2, STATICCALL, PUSH1, retSize,
			PUSH1, retOff, RETURN)
		caller := makeAccountWithCode(t, st, "caller", callerCode)

		addToBalance(t, st, callee, 100000)
		txe := runVM(st, caller, callee, callerCode, 1000)
		require.NotNil(t, txe.Exception)
		require.Equal(t, errors.Codes.IllegalWrite, txe.Exception.ErrorCode(),
			"expected static call violation because of call with value")
	})

	t.Run("StaticCallNoValue", func(t *testing.T) {
		gas1, gas2 := byte(0x1), byte(0x1)
		value := byte(0x69)
		var inOff, inSize, retOff, retSize byte

		// this final test just checks that STATICCALL actually works
		st := acmstate.NewMemoryState()

		finalAddress := makeAccountWithCode(t, st, "final", MustSplice(PUSH1, int64(20), return1()))
		// intermediate account CALLs another contract *without* a value
		callee := makeAccountWithCode(t, st, "callee", MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1,
			inOff, PUSH1, 0x00, PUSH20, finalAddress, PUSH2, gas1, gas2, CALL, returnWord()))

		callerCode := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1,
			inOff, PUSH1, value, PUSH20, callee, PUSH2, gas1, gas2, STATICCALL, PUSH1, retSize,
			PUSH1, retOff, RETURN)
		caller := makeAccountWithCode(t, st, "caller", callerCode)

		addToBalance(t, st, callee, 100000)
		txe := runVM(st, caller, callee, callerCode, 1000)
		// no exceptions expected because value never set in children
		require.NoError(t, txe.Exception.AsError())
		exCalls := txe.ExceptionalCalls()
		require.Len(t, exCalls, 0)
	})

	// Test evm account creation
	t.Run("Create", func(t *testing.T) {
		st := acmstate.NewMemoryState()

		code := MustSplice(PUSH1, 0x0, PUSH1, 0x0, PUSH1, 0x0, CREATE, PUSH1, 0, MSTORE, PUSH1, 20, PUSH1, 12, RETURN)
		callee := makeAccountWithCode(t, st, "callee", code)
		// ensure pre-generated address has same sequence number
		nonce := make([]byte, txs.HashLength+uint64Length)
		binary.BigEndian.PutUint64(nonce[txs.HashLength:], 1)
		addr := crypto.NewContractAddress(callee, nonce)

		var gas uint64 = 100000
		caller := newAccount(t, st, "1, 2, 3")
		output, err := call(vm, st, caller, callee, code, nil, &gas)
		assert.NoError(t, err, "Should return new address without error")
		assert.Equal(t, addr.Bytes(), output, "Addresses should be equal")
	})

	// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1014.md
	t.Run("Create2", func(t *testing.T) {
		st := acmstate.NewMemoryState()

		// salt of 0s
		var salt [32]byte
		code := MustSplice(PUSH1, 0x0, PUSH1, 0x0, PUSH1, 0x0, PUSH32, salt[:], CREATE2, PUSH1, 0, MSTORE, PUSH1, 20, PUSH1, 12, RETURN)
		callee := makeAccountWithCode(t, st, "callee", code)
		addr := crypto.NewContractAddress2(callee, salt, code)

		var gas uint64 = 100000
		caller := newAccount(t, st, "1, 2, 3")
		output, err := call(vm, st, caller, callee, code, nil, &gas)
		assert.NoError(t, err, "Should return new address without error")
		assert.Equal(t, addr.Bytes(), output, "Returned value not equal to create2 address")
	})

	// This test was introduced to cover an issues exposed in our handling of the
	// gas limit passed from caller to callee on various forms of CALL.
	// The idea of this test is to implement a simple DelegateCall in EVM code
	// We first run the DELEGATECALL with _just_ enough gas expecting a simple return,
	// and then run it with 1 gas unit less, expecting a failure
	t.Run("DelegateCallGas", func(t *testing.T) {
		st := acmstate.NewMemoryState()

		inOff := 0
		inSize := 0 // no call data
		retOff := 0
		retSize := 32
		calleeReturnValue := int64(20)

		callee := makeAccountWithCode(t, st, "callee", MustSplice(PUSH1, calleeReturnValue, PUSH1, 0, MSTORE, PUSH1, 32, PUSH1, 0, RETURN))

		// 6 op codes total
		baseOpsCost := native.GasBaseOp * 6
		// 4 pushes
		pushCost := native.GasStackOp * 4
		// 2 pushes 2 pops
		returnCost := native.GasStackOp * 4
		// To push success/failure
		resumeCost := native.GasStackOp

		// Gas is not allowed to drop to 0 so we add resumecost
		delegateCallCost := baseOpsCost + pushCost + returnCost + resumeCost

		// Here we split up the caller code so we can make a DELEGATE call with
		// different amounts of gas. The value we sandwich in the middle is the amount
		// we subtract from the available gas (that the caller has available), so:
		// code := MustSplice(callerCodePrefix, <amount to subtract from GAS> , callerCodeSuffix)
		// gives us the code to make the call
		callerCodePrefix := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize,
			PUSH1, inOff, PUSH20, callee, PUSH1)
		callerCodeSuffix := MustSplice(DELEGATECALL, returnWord())

		// Perform a delegate call
		//callerCode := MustSplice(callerCodePrefix,
		//	Give just enough gas to make the DELEGATECALL
		//delegateCallCost, callerCodeSuffix)
		//caller := makeAccountWithCode(t, st, "caller", callerCode)

		// Should pass
		//txe := runVM(st, caller, callee, callerCode, 100)
		//assert.Nil(t, txe.Exception, "Should have sufficient funds for call")
		//assert.Equal(t, Int64ToWord256(calleeReturnValue).Bytes(), txe.Result.Return)

		callerCode2 := MustSplice(callerCodePrefix,
			// Shouldn't be enough gas to make call
			delegateCallCost-1, callerCodeSuffix)
		caller2 := makeAccountWithCode(t, st, "caller2", callerCode2)

		// Should fail
		txe := runVM(st, caller2, callee, callerCode2, 100)
		assert.NotNil(t, txe.Exception, "Should have insufficient gas for call")
	})

	t.Run("MemoryBounds", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		blockchain := new(blockchain)
		eventSink := exec.NewNoopEventSink()
		vm := New(Options{
			MemoryProvider: func(err errors.Sink) Memory {
				return NewDynamicMemory(1024, 2048, err)
			},
		})
		caller := makeAccountWithCode(t, st, "caller", nil)
		callee := makeAccountWithCode(t, st, "callee", nil)
		gas := uint64(100000)
		word := One256
		// This attempts to store a value at the memory boundary and return it
		params := engine.CallParams{
			Gas:    &gas,
			Caller: caller,
			Callee: callee,
		}
		code := MustSplice(pushWord(word), storeAtEnd(), MLOAD, storeAtEnd(), returnAfterStore())

		output, err := vm.Execute(st, blockchain, eventSink, params, code)
		assert.NoError(t, err)
		assert.Equal(t, word.Bytes(), output)

		// Same with number
		word = Int64ToWord256(232234234432)
		code = MustSplice(pushWord(word), storeAtEnd(), MLOAD, storeAtEnd(), returnAfterStore())
		output, err = vm.Execute(st, blockchain, eventSink, params, code)
		assert.NoError(t, err)
		assert.Equal(t, word.Bytes(), output)

		// Now test a series of boundary stores
		code = pushWord(word)
		for i := 0; i < 10; i++ {
			code = MustSplice(code, storeAtEnd(), MLOAD)
		}
		code = MustSplice(code, storeAtEnd(), returnAfterStore())
		output, err = vm.Execute(st, blockchain, eventSink, params, code)
		assert.NoError(t, err)
		assert.Equal(t, word.Bytes(), output)

		// Same as above but we should breach the upper memory limit set in memoryProvider
		code = pushWord(word)
		for i := 0; i < 100; i++ {
			code = MustSplice(code, storeAtEnd(), MLOAD)
		}
		code = MustSplice(code, storeAtEnd(), returnAfterStore())
		_, err = vm.Execute(st, blockchain, eventSink, params, code)
		assert.Error(t, err, "Should hit memory out of bounds")
	})

	t.Run("MsgSender", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		account1 := newAccount(t, st, "1, 2, 3")
		account2 := newAccount(t, st, "3, 2, 1")
		var gas uint64 = 100000

		/*
				pragma solidity ^0.5.4;

				contract SimpleStorage {
			                function get() public constant returns (address) {
			        	        return msg.sender;
			    	        }
				}
		*/

		// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
		code := hex.MustDecodeString("6060604052341561000f57600080fd5b60ca8061001d6000396000f30060606040526004361060" +
			"3f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680636d4ce63c14604457" +
			"5b600080fd5b3415604e57600080fd5b60546096565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ff" +
			"ffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6000339050905600a165627a" +
			"7a72305820b9ebf49535372094ae88f56d9ad18f2a79c146c8f56e7ef33b9402924045071e0029")

		var err error
		// Run the contract initialisation code to obtain the contract code that would be mounted at account2
		contractCode, err := call(vm, st, account1, account2, code, code, &gas)
		require.NoError(t, err)

		// Not needed for this test (since contract code is passed as argument to vm), but this is what an execution
		// framework must do
		err = native.InitCode(st, account2, contractCode)
		require.NoError(t, err)

		// Input is the function hash of `get()`
		input := hex.MustDecodeString("6d4ce63c")

		output, err := call(vm, st, account1, account2, contractCode, input, &gas)
		require.NoError(t, err)

		assert.Equal(t, account1.Word256().Bytes(), output)
	})

	t.Run("Invalid", func(t *testing.T) {
		st := acmstate.NewMemoryState()

		// Create accounts
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "1, 0, 1")

		var gas uint64 = 100000

		bytecode := MustSplice(PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61,
			0x67, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, INVALID)

		output, err := call(vm, st, account1, account2, bytecode, nil, &gas)
		assert.Equal(t, errors.Codes.ExecutionAborted, errors.GetCode(err))
		t.Logf("Output: %v Error: %v\n", output, err)
	})

	t.Run("ReturnDataSize", func(t *testing.T) {
		st := acmstate.NewMemoryState()

		accountName := "account2addresstests"

		ret := "My return message"
		callcode := MustSplice(PUSH32, RightPadWord256([]byte(ret)), PUSH1, 0x00, MSTORE, PUSH1, len(ret), PUSH1, 0x00, RETURN)

		// Create accounts
		account1 := newAccount(t, st, "1")
		account2 := makeAccountWithCode(t, st, accountName, callcode)

		var gas uint64 = 100000

		gas1, gas2 := byte(0x1), byte(0x1)
		value := byte(0x69)
		inOff, inSize := byte(0x0), byte(0x0) // no call data
		retOff, retSize := byte(0x0), byte(len(ret))

		bytecode := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1, inOff, PUSH1, value,
			PUSH20, account2, PUSH2, gas1, gas2, CALL,
			RETURNDATASIZE, PUSH1, 0x00, MSTORE, PUSH1, 0x20, PUSH1, 0x00, RETURN)

		expected := Uint64ToWord256(uint64(len(ret))).Bytes()

		output, err := call(vm, st, account1, account2, bytecode, nil, &gas)
		require.NoError(t, err)
		assert.Equal(t, expected, output)

		t.Logf("Output: %v Error: %v\n", output, err)

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ReturnDataCopy", func(t *testing.T) {
		st := acmstate.NewMemoryState()

		accountName := "account2addresstests"

		ret := "My return message"
		callcode := MustSplice(PUSH32, RightPadWord256([]byte(ret)), PUSH1, 0x00, MSTORE, PUSH1, len(ret), PUSH1, 0x00, RETURN)

		// Create accounts
		account1 := newAccount(t, st, "1")
		account2 := makeAccountWithCode(t, st, accountName, callcode)

		var gas uint64 = 100000

		gas1, gas2 := byte(0x1), byte(0x1)
		value := byte(0x69)
		inOff, inSize := byte(0x0), byte(0x0) // no call data
		retOff, retSize := byte(0x0), byte(len(ret))

		bytecode := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1, inOff, PUSH1, value,
			PUSH20, account2, PUSH2, gas1, gas2, CALL, RETURNDATASIZE, PUSH1, 0x00, PUSH1, 0x00, RETURNDATACOPY,
			RETURNDATASIZE, PUSH1, 0x00, RETURN)

		expected := []byte(ret)

		output, err := call(vm, st, account1, account2, bytecode, nil, &gas)
		require.NoError(t, err)
		assert.Equal(t, expected, output)

		t.Logf("Output: %v Error: %v\n", output, err)
	})

	t.Run("CallNonExistent", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		blockchain := new(blockchain)
		eventSink := exec.NewNoopEventSink()
		account1 := newAccount(t, st, "1")
		addToBalance(t, st, account1, 100000)
		unknownAddress := native.AddressFromName("nonexistent")
		var gas uint64
		amt := uint64(100)
		params := engine.CallParams{
			Caller: account1,
			Callee: unknownAddress,
			Value:  amt,
			Gas:    &gas,
		}
		_, ex := vm.Execute(st, blockchain, eventSink, params, nil)
		require.Equal(t, errors.Codes.NonExistentAccount, errors.GetCode(ex),
			"Should not be able to call account before creating it (even before initialising)")
		acc, err := st.GetAccount(unknownAddress)
		require.NoError(t, err)
		require.Nil(t, acc)
	})

	t.Run("GetBlockHash", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		blockchain := new(blockchain)
		eventSink := exec.NewNoopEventSink()

		// Create accounts
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "101")

		var gas uint64 = 100000

		bytecode := MustSplice(PUSH1, 2, BLOCKHASH)

		params := engine.CallParams{
			Caller: account1,
			Callee: account2,
			Gas:    &gas,
		}
		// Non existing block
		blockchain.blockHeight = 1
		_, err := vm.Execute(st, blockchain, eventSink, params, bytecode)
		require.Equal(t, errors.Codes.InvalidBlockNumber, errors.GetCode(err),
			"attempt to get block hash of a non-existent block")

		// Excessive old block
		blockchain.blockHeight = 258
		bytecode = MustSplice(PUSH1, 1, BLOCKHASH)

		_, err = vm.Execute(st, blockchain, eventSink, params, bytecode)
		require.Equal(t, errors.Codes.BlockNumberOutOfRange, errors.GetCode(err),
			"attempt to get block hash of a block outside of allowed range")

		// Get block hash
		blockchain.blockHeight = 257
		bytecode = MustSplice(PUSH1, 2, BLOCKHASH, return1())

		output, err := vm.Execute(st, blockchain, eventSink, params, bytecode)
		assert.NoError(t, err)
		assert.Equal(t, LeftPadWord256([]byte{2}), LeftPadWord256(output))

		// Get block hash fail
		blockchain.blockHeight = 3
		bytecode = MustSplice(PUSH1, 4, BLOCKHASH, return1())

		_, err = vm.Execute(st, blockchain, eventSink, params, bytecode)
		require.Equal(t, errors.Codes.InvalidBlockNumber, errors.GetCode(err),
			"attempt to get block hash failed")
	})

	t.Run("PushWord", func(t *testing.T) {
		word := Int64ToWord256(int64(2133213213))
		assert.Equal(t, MustSplice(PUSH4, 0x7F, 0x26, 0x40, 0x1D), pushWord(word))
		word[0] = 1
		assert.Equal(t, MustSplice(PUSH32,
			1, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0x7F, 0x26, 0x40, 0x1D), pushWord(word))
		assert.Equal(t, MustSplice(PUSH1, 0), pushWord(Word256{}))
		assert.Equal(t, MustSplice(PUSH1, 1), pushWord(Int64ToWord256(1)))
	})

	// Kind of indirect test of Splice, but here to avoid import cycles
	t.Run("Bytecode", func(t *testing.T) {
		assert.Equal(t,
			MustSplice(1, 2, 3, 4, 5, 6),
			MustSplice(1, 2, 3, MustSplice(4, 5, 6)))
		assert.Equal(t,
			MustSplice(1, 2, 3, 4, 5, 6, 7, 8),
			MustSplice(1, 2, 3, MustSplice(4, MustSplice(5), 6), 7, 8))
		assert.Equal(t,
			MustSplice(PUSH1, 2),
			MustSplice(byte(PUSH1), 0x02))
		assert.Equal(t,
			[]byte{},
			MustSplice(MustSplice(MustSplice())))

		contractAccount := &acm.Account{Address: crypto.AddressFromWord256(Int64ToWord256(102))}
		addr := contractAccount.Address
		gas1, gas2 := byte(0x1), byte(0x1)
		value := byte(0x69)
		inOff, inSize := byte(0x0), byte(0x0) // no call data
		retOff, retSize := byte(0x0), byte(0x20)
		contractCodeBytecode := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1,
			inOff, PUSH1, value, PUSH20, addr, PUSH2, gas1, gas2, CALL, PUSH1, retSize,
			PUSH1, retOff, RETURN)
		contractCode := []byte{0x60, retSize, 0x60, retOff, 0x60, inSize, 0x60, inOff, 0x60, value, 0x73}
		contractCode = append(contractCode, addr[:]...)
		contractCode = append(contractCode, []byte{0x61, gas1, gas2, 0xf1, 0x60, 0x20, 0x60, 0x0, 0xf3}...)
		assert.Equal(t, contractCode, contractCodeBytecode)
	})

	t.Run("Concat", func(t *testing.T) {
		assert.Equal(t,
			[]byte{0x01, 0x02, 0x03, 0x04},
			Concat([]byte{0x01, 0x02}, []byte{0x03, 0x04}))
	})

	t.Run("Subslice", func(t *testing.T) {
		const size = 10
		data := make([]byte, size)
		for i := 0; i < size; i++ {
			data[i] = byte(i)
		}
		for n := uint64(0); n < size; n++ {
			data = data[:n]
			for offset := uint64(0); offset < size; offset++ {
				for length := uint64(0); length < size; length++ {
					_, err := subslice(data, offset, length)
					if offset < 0 || length < 0 || n < offset {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				}
			}
		}

		bs, err := subslice([]byte{1, 2, 3, 4, 5, 6, 7, 8}, 4, 32)
		require.NoError(t, err)
		assert.Equal(t, []byte{
			5, 6, 7, 8, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
		}, bs)
	})

	t.Run("DataStackOverflow", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		blockchain := new(blockchain)
		eventSink := exec.NewNoopEventSink()
		account1 := newAccount(t, st, "1, 2, 3")
		account2 := newAccount(t, st, "3, 2, 1")

		var gas uint64 = 100000

		/*
			pragma solidity ^0.5.4;

			contract SimpleStorage {
				function get() public constant returns (address) {
					return get();
				}
			}
		*/

		// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
		code, err := hex.DecodeString("608060405234801561001057600080fd5b5060d18061001f6000396000f300608060405260043610" +
			"603f576000357c0100000000000000000000000000000000000000000000000000000000900463ff" +
			"ffffff1680636d4ce63c146044575b600080fd5b348015604f57600080fd5b5060566098565b6040" +
			"51808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffff" +
			"ffffffffffff16815260200191505060405180910390f35b600060a06098565b9050905600a16562" +
			"7a7a72305820daacfba0c21afacb5b67f26bc8021de63eaa560db82f98357d4e513f3249cf350029")
		require.NoError(t, err)

		// Run the contract initialisation code to obtain the contract code that would be mounted at account2
		params := engine.CallParams{
			Caller: account1,
			Callee: account2,
			Input:  code,
			Gas:    &gas,
		}
		vm := New(Options{
			DataStackMaxDepth: 4,
		})

		code, err = vm.Execute(st, blockchain, eventSink, params, code)
		require.NoError(t, err)

		// Input is the function hash of `get()`
		params.Input, err = hex.DecodeString("6d4ce63c")
		require.NoError(t, err)

		_, ex := vm.Execute(st, blockchain, eventSink, params, code)
		require.Equal(t, errors.Codes.DataStackOverflow, errors.GetCode(ex), "Should be stack overflow")
	})

	t.Run("CallStackOverflow", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		blockchain := new(blockchain)
		eventSink := exec.NewNoopEventSink()
		txe := new(exec.TxExecution)

		account1 := newAccount(t, st, "1, 2, 3")
		account2 := newAccount(t, st, "3, 2, 1")

		// Sender accepts lot of gaz but we run on a caped call stack node
		var gas uint64 = 100000
		/*
			pragma solidity ^0.5.4;

			contract A {
			   function callMeBack() public {
					return require(msg.sender.call(bytes4(keccak256("callMeBack()")),this));
				}
			}
		*/

		// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
		code, err := hex.DecodeString("608060405234801561001057600080fd5b5061017a806100206000396000f3006080604052600436" +
			"10610041576000357c01000000000000000000000000000000000000000000000000000000009004" +
			"63ffffffff168063692c3b7c14610046575b600080fd5b34801561005257600080fd5b5061005b61" +
			"005d565b005b3373ffffffffffffffffffffffffffffffffffffffff1660405180807f63616c6c4d" +
			"654261636b28290000000000000000000000000000000000000000815250600c0190506040518091" +
			"0390207c010000000000000000000000000000000000000000000000000000000090043060405182" +
			"63ffffffff167c010000000000000000000000000000000000000000000000000000000002815260" +
			"0401808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffff" +
			"ffffffffffffff1681526020019150506000604051808303816000875af192505050151561014c57" +
			"600080fd5b5600a165627a7a723058209315a40abb8b23b7c2a340e938b01367b419a23818475a2e" +
			"ee80d09da3f7ba780029")
		require.NoError(t, err)

		params := engine.CallParams{
			Caller: account1,
			Callee: account2,
			Input:  code,
			Gas:    &gas,
		}
		options := Options{
			CallStackMaxDepth: 2,
		}
		vm := New(options)
		// Run the contract initialisation code to obtain the contract code that would be mounted at account2
		contractCode, err := vm.Execute(st, blockchain, eventSink, params, code)
		require.NoError(t, err)

		err = native.InitCode(st, account1, contractCode)
		require.NoError(t, err)
		err = native.InitCode(st, account2, contractCode)
		require.NoError(t, err)

		// keccak256 hash of 'callMeBack()'
		params.Input, err = hex.DecodeString("692c3b7c")
		require.NoError(t, err)

		_, err = vm.Execute(st, blockchain, txe, params, contractCode)
		// The TxExecution must be an exception to get the callerror
		txe.PushError(err)
		require.Error(t, err)
		callError := txe.CallError()
		require.Error(t, callError)
		require.Equal(t, errors.Codes.ExecutionReverted, errors.GetCode(callError))
		// Errors are post-order so first is deepest
		require.True(t, len(callError.NestedErrors) > 0)
		deepestErr := callError.NestedErrors[0]
		require.Equal(t, errors.Codes.CallStackOverflow, errors.GetCode(deepestErr))
		assert.Equal(t, options.CallStackMaxDepth, deepestErr.StackDepth)
		assert.Equal(t, account2, deepestErr.Callee)
		assert.Equal(t, account1, deepestErr.Caller)
	})

	t.Run("ExtCodeHash", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "101")

		var gas uint64 = 100000

		// The EXTCODEHASH of the account without code is c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
		//    what is the keccack256 hash of empty data.
		bytecode := MustSplice(PUSH20, account1, EXTCODEHASH, return1())
		output, err := call(vm, st, account1, account2, bytecode, nil, &gas)
		assert.NoError(t, err)
		assert.Equal(t, hex.MustDecodeString("c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"), output)

		// The EXTCODEHASH of a native account is hash of its name.
		bytecode = MustSplice(PUSH1, 0x03, EXTCODEHASH, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		assert.NoError(t, err)
		assert.Equal(t, crypto.Keccak256([]byte("ripemd160Func")), output)

		// EXTCODEHASH of non-existent account should be 0
		bytecode = MustSplice(PUSH1, 0xff, EXTCODEHASH, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		assert.NoError(t, err)
		assert.Equal(t, Zero256[:], output)

		// EXTCODEHASH is the hash of an account code
		acc := makeAccountWithCode(t, st, "trustedCode", MustSplice(BLOCKHASH, return1()))
		bytecode = MustSplice(PUSH20, acc, EXTCODEHASH, return1())
		output, err = call(vm, st, account1, account2, bytecode, nil, &gas)
		assert.NoError(t, err)
		assert.Equal(t, hex.MustDecodeString("010da270094b5199d3e54f89afe4c66cdd658dd8111a41998714227e14e171bd"), output)
	})

	// Tests logs and events.
	t.Run("TestLogEvents", func(t *testing.T) {
		expectedData := []byte{0x10}
		expectedHeight := uint64(0)
		expectedTopics := []Word256{
			Int64ToWord256(1),
			Int64ToWord256(2),
			Int64ToWord256(3),
			Int64ToWord256(4)}

		st := acmstate.NewMemoryState()
		blockchain := new(blockchain)
		// Create accounts
		address1 := crypto.Address{1, 3, 5, 7, 9}
		address2 := crypto.Address{2, 4, 6, 8, 10}

		err := native.CreateAccount(st, address1)
		require.NoError(t, err)
		err = native.CreateAccount(st, address2)
		require.NoError(t, err)

		var gas uint64 = 100000

		mstore8 := byte(MSTORE8)
		push1 := byte(PUSH1)
		log4 := byte(LOG4)
		stop := byte(STOP)

		code := []byte{
			push1, 16, // data value
			push1, 0, // memory slot
			mstore8,
			push1, 4, // topic 4
			push1, 3, // topic 3
			push1, 2, // topic 2
			push1, 1, // topic 1
			push1, 1, // size of data
			push1, 0, // data starts at this offset
			log4,
			stop,
		}

		txe := new(exec.TxExecution)

		params := engine.CallParams{
			Caller: address1,
			Callee: address2,
			Gas:    &gas,
		}
		_, err = vm.Execute(st, blockchain, txe, params, code)
		require.NoError(t, err)

		for _, ev := range txe.Events {
			if ev.Log != nil {
				if !reflect.DeepEqual(ev.Log.Topics, expectedTopics) {
					t.Errorf("Event topics are wrong. Got: %v. Expected: %v", ev.Log.Topics, expectedTopics)
				}
				if !bytes.Equal(ev.Log.Data, expectedData) {
					t.Errorf("Event data is wrong. Got: %s. Expected: %s", ev.Log.Data, expectedData)
				}
				if ev.Header.Height != expectedHeight {
					t.Errorf("Event block height is wrong. Got: %d. Expected: %d", ev.Header.Height, expectedHeight)
				}
				return
			}
		}
		t.Fatalf("Did not see LogEvent")
	})

	t.Run("BigModExp", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "101")

		// The solidity compiled contract. It calls bigmodexp with b,e,m inputs and compares the result with proof, where m is the mod, b the base, e the exp, and proof the expected result.
		bytecode := solidity.DeployedBytecode_BigMod

		// The function "expmod" is an assertion. It takes the base, exponent, modulus, and the expected value and
		// returns 1 if the values match.
		spec, err := abi.ReadSpec(solidity.Abi_BigMod)
		require.NoError(t, err)

		expModFunctionID := spec.Functions["expmod"].FunctionID

		n := int64(10)
		for base := -n; base < n; base++ {
			for exp := -n; exp < n; exp++ {
				for mod := int64(1); mod < n; mod++ {
					b := big.NewInt(base)
					e := big.NewInt(exp)
					m := big.NewInt(mod)
					v := new(big.Int).Exp(b, e, m)
					if v == nil {
						continue
					}

					input := MustSplice(expModFunctionID, // expmod function
						BigIntToWord256(b), BigIntToWord256(e), BigIntToWord256(m), // base^exp % mod
						BigIntToWord256(v)) // == expected

					gas := uint64(10000000)
					out, err := call(vm, st, account1, account2, bytecode, input, &gas)

					require.NoError(t, err)

					require.Equal(t, One256, LeftPadWord256(out), "expected %d^%d mod %d == %d",
						base, exp, mod, e)
				}
			}
		}
	})

	t.Run("SDIV", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		blockchain := new(blockchain)
		eventSink := exec.NewNoopEventSink()

		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "101")
		var gas uint64 = 100000

		tests := []struct {
			name     string
			bytecode []byte
			expected []byte
		}{
			{
				// (SDIV 7 3 )
				name: "sdiv0",
				bytecode: MustSplice(PUSH1, 0x03, PUSH1, 0x07, SDIV, return1()),
				expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
			},
			{
				// (SDIV 7 (- 0 3) )
				name: "sdiv1",
				bytecode: MustSplice(PUSH1, 0x03, PUSH1, 0x00, SUB, PUSH1, 0x07, SDIV, return1()),
				expected: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE},
			},
			{
				// (SDIV (- 0 7) 3 )
				name: "sdiv2",
				bytecode: MustSplice(PUSH1, 0x03, PUSH1, 0x07, PUSH1, 0x00, SUB, SDIV, return1()),
				expected: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE},
			},
			{
				// (SDIV (- 0 7) (- 0 3) )
				name: "sdiv3",
				bytecode: MustSplice(PUSH1, 0x03, PUSH1, 0x00, SUB, PUSH1, 0x07, PUSH1, 0x00, SUB, SDIV, return1()),
				expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
			},
		}
		for _, tt := range tests {
			params := engine.CallParams{
				Caller: account1,
				Callee: account2,
				Gas:    &gas,
			}
			if output, err := vm.Execute(st, blockchain, eventSink, params, tt.bytecode); err != nil || !bytes.Equal(output, tt.expected) {
				t.Errorf("Reported error in %v.", tt.name)
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, output)
			}
		}
	})

	t.Run("SMOD", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		blockchain := new(blockchain)
		eventSink := exec.NewNoopEventSink()

		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "101")
		var gas uint64 = 100000

		tests := []struct {
			name     string
			bytecode []byte
			expected []byte
		}{
			{
				// (SMOD 7 3)
				name: "smod0",
				bytecode: MustSplice(PUSH1, 0x03, PUSH1, 0x07, SMOD, return1()),
				expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			},
			{
				// (SMOD 7 (- 0 3))
				name: "smod1",
				bytecode: MustSplice(PUSH1, 0x03, PUSH1, 0x00, SUB, PUSH1, 0x07, SMOD, return1()),
				expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			},
			{
				// (SMOD (- 0 7) 3)
				name: "smod2",
				bytecode: MustSplice(PUSH1, 0x03, PUSH1, 0x07, PUSH1, 0x00, SUB, SMOD, return1()),
				expected: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			},
			{
				// (SMOD (- 0 7) (- 0 3))
				name: "smod3",
				bytecode: MustSplice(PUSH1, 0x03, PUSH1, 0x00, SUB, PUSH1, 0x07, PUSH1, 0x00, SUB, SMOD, return1()),
				expected: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			},
		}
		for _, tt := range tests {
			params := engine.CallParams{
				Caller: account1,
				Callee: account2,
				Gas:    &gas,
			}
			if output, err := vm.Execute(st, blockchain, eventSink, params, tt.bytecode); err != nil || !bytes.Equal(output, tt.expected) {
				t.Errorf("Reported error in %v.", tt.name)
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, output)
			}
		}
	})

	t.Run("SIGNEXTEND", func(t *testing.T) {
		st := acmstate.NewMemoryState()
		blockchain := new(blockchain)
		eventSink := exec.NewNoopEventSink()

		account1 := newAccount(t, st, "1")
		account2 := newAccount(t, st, "101")
		var gas uint64 = 100000

		tests := []struct {
			name     string
			bytecode []byte
			expected []byte
		}{
			{
				// 0x6A == 0b01101010
				name: "signextend0",
				bytecode: MustSplice(PUSH3, 0x12, 0x2F, 0x6A, PUSH1, 0x00, SIGNEXTEND, return1()),
				expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6A},
			},
			{
				name: "signextend1",
				bytecode: MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, SIGNEXTEND, return1()),
				expected: []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
					0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			},
		}
		for _, tt := range tests {
			params := engine.CallParams{
				Caller: account1,
				Callee: account2,
				Gas:    &gas,
			}
			if output, err := vm.Execute(st, blockchain, eventSink, params, tt.bytecode); err != nil || !bytes.Equal(output, tt.expected) {
				t.Errorf("Reported error in %v.", tt.name)
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, output)
			}
		}
	})
}

type blockchain struct {
	blockHeight uint64
	blockTime   time.Time
}

func (b *blockchain) LastBlockHeight() uint64 {
	return b.blockHeight
}

func (b *blockchain) LastBlockTime() time.Time {
	return b.blockTime
}

func (b *blockchain) BlockHash(height uint64) ([]byte, error) {
	if height > b.blockHeight {
		return nil, errors.Codes.InvalidBlockNumber
	}
	bs := make([]byte, 32)
	binary.BigEndian.PutUint64(bs[24:], height)
	return bs, nil
}

// helpers

func newAccount(t testing.TB, st acmstate.ReaderWriter, name string) crypto.Address {
	address := native.AddressFromName(name)
	err := native.CreateAccount(st, address)
	require.NoError(t, err)
	return address
}

func makeAccountWithCode(t testing.TB, st acmstate.ReaderWriter, name string, code []byte) crypto.Address {
	address := native.AddressFromName(name)
	err := native.CreateAccount(st, address)
	require.NoError(t, err)
	err = native.InitCode(st, address, code)
	require.NoError(t, err)
	addToBalance(t, st, address, 9999999)
	return address
}

func addToBalance(t testing.TB, st acmstate.ReaderWriter, address crypto.Address, amount uint64) {
	err := native.UpdateAccount(st, address, func(account *acm.Account) error {
		return account.AddToBalance(amount)
	})
	require.NoError(t, err)
}

func call(vm *EVM, st acmstate.ReaderWriter, origin, callee crypto.Address, code []byte, input []byte,
	gas *uint64) ([]byte, error) {

	evs := new(exec.Events)
	out, err := vm.Execute(st, new(blockchain), evs, engine.CallParams{
		Caller: origin,
		Callee: callee,
		Input:  input,
		Gas:    gas,
	}, code)

	if err != nil {
		return nil, &errors.CallError{
			CodedError:   errors.AsException(err),
			NestedErrors: evs.NestedCallErrors(),
		}
	}
	return out, nil
}

// These code segment helpers exercise the MSTORE MLOAD MSTORE cycle to test
// both of the memory operations. Each MSTORE is done on the memory boundary
// (at MSIZE) which Solidity uses to find guaranteed unallocated memory.

// storeAtEnd expects the value to be stored to be on top of the stack, it then
// stores that value at the current memory boundary
func storeAtEnd() []byte {
	// Pull in MSIZE (to carry forward to MLOAD), swap in value to store, store it at MSIZE
	return MustSplice(MSIZE, SWAP1, DUP2, MSTORE)
}

func returnAfterStore() []byte {
	return MustSplice(PUSH1, 32, DUP2, RETURN)
}

// Store the top element of the stack (which is a 32-byte word) in memory
// and return it. Useful for a simple return value.
func return1() []byte {
	return MustSplice(PUSH1, 0, MSTORE, returnWord())
}

func returnWord() []byte {
	// PUSH1 => return size, PUSH1 => return offset, RETURN
	return MustSplice(PUSH1, 32, PUSH1, 0, RETURN)
}

// Subscribes to an AccCall, runs the vm, returns the output any direct exception
// and then waits for any exceptions transmitted by Data in the AccCall
// event (in the case of no direct error from call we will block waiting for
// at least 1 AccCall event)
func runVM(st acmstate.ReaderWriter, caller, callee crypto.Address, code []byte, gas uint64) *exec.TxExecution {
	gasBefore := gas
	txe := new(exec.TxExecution)
	vm := New(Options{
		DebugOpcodes: true,
	})
	params := engine.CallParams{
		Caller: caller,
		Callee: callee,
		Gas:    &gas,
	}
	output, err := vm.Execute(st, new(blockchain), txe, params, code)
	txe.PushError(err)
	for _, ev := range txe.ExceptionalCalls() {
		txe.PushError(ev.Header.Exception)
	}
	txe.Return(output, gasBefore-gas)
	return txe
}

// this is code to call another contract (hardcoded as addr)
func callContractCode(addr crypto.Address) []byte {
	gas1, gas2 := byte(0x1), byte(0x1)
	value := byte(0x69)
	inOff, inSize := byte(0x0), byte(0x0) // no call data
	retOff, retSize := byte(0x0), byte(0x20)
	// this is the code we want to run (send funds to an account and return)
	return MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1,
		inOff, PUSH1, value, PUSH20, addr, PUSH2, gas1, gas2, CALL, PUSH1, retSize,
		PUSH1, retOff, RETURN)
}

// Produce bytecode for a PUSH<N>, b_1, ..., b_N where the N is number of bytes
// contained in the unpadded word
func pushWord(word Word256) []byte {
	leadingZeros := byte(0)
	for leadingZeros < 32 {
		if word[leadingZeros] == 0 {
			leadingZeros++
		} else {
			return MustSplice(byte(PUSH32)-leadingZeros, word[leadingZeros:])
		}
	}
	return MustSplice(PUSH1, 0)
}
