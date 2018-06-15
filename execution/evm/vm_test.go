// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package evm

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/errors"
	. "github.com/hyperledger/burrow/execution/evm/asm"
	. "github.com/hyperledger/burrow/execution/evm/asm/bc"
	evm_events "github.com/hyperledger/burrow/execution/evm/events"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ripemd160"
)

// Test output is a bit clearer if we /dev/null the logging, but can be re-enabled by uncommenting the below
//var logger, _, _ = lifecycle.NewStdErrLogger()
//
var logger = logging.NewNoopLogger()

func newAppState() *FakeAppState {
	fas := &FakeAppState{
		accounts: make(map[crypto.Address]acm.Account),
		storage:  make(map[string]Word256),
	}
	// For default permissions
	fas.accounts[acm.GlobalPermissionsAddress] = acm.ConcreteAccount{
		Permissions: permission.DefaultAccountPermissions,
	}.Account()
	return fas
}

func newParams() Params {
	return Params{
		BlockHeight: 0,
		BlockHash:   Zero256,
		BlockTime:   0,
		GasLimit:    0,
	}
}

func newAccount(seed ...byte) *acm.MutableAccount {
	hasher := ripemd160.New()
	hasher.Write(seed)
	return acm.ConcreteAccount{
		Address: crypto.MustAddressFromBytes(hasher.Sum(nil)),
	}.MutableAccount()
}

// Runs a basic loop
func TestVM(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)

	// Create accounts
	account1 := newAccount(1)
	account2 := newAccount(1, 0, 1)

	var gas uint64 = 100000

	bytecode := MustSplice(PUSH1, 0x00, PUSH1, 0x20, MSTORE, JUMPDEST, PUSH2, 0x0F, 0x0F, PUSH1, 0x20, MLOAD,
		SLT, ISZERO, PUSH1, 0x1D, JUMPI, PUSH1, 0x01, PUSH1, 0x20, MLOAD, ADD, PUSH1, 0x20,
		MSTORE, PUSH1, 0x05, JUMP, JUMPDEST)

	start := time.Now()
	output, err := ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	fmt.Printf("Output: %v Error: %v\n", output, err)
	fmt.Println("Call took:", time.Since(start))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSHL(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)
	account1 := newAccount(1)
	account2 := newAccount(1, 0, 1)

	var gas uint64 = 100000

	//Shift left 0
	bytecode := MustSplice(PUSH1, 0x01, PUSH1, 0x00, SHL, return1())
	output, err := ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value := []uint8([]byte{0x1})
	expected := LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift left 0
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x00, SHL, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Shift left 1
	bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0x01, SHL, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x2})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift left 1
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SHL, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift left 1
	bytecode = MustSplice(PUSH32, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SHL, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Shift left 255
	bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0xFF, SHL, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x80})
	expected = RightPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift left 255
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SHL, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x80})
	expected = RightPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Shift left 256 (overflow)
	bytecode = MustSplice(PUSH1, 0x01, PUSH2, 0x01, 0x00, SHL, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x00})
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
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Shift left 257 (overflow)
	bytecode = MustSplice(PUSH1, 0x01, PUSH2, 0x01, 0x01, SHL, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

}

func TestSHR(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)
	account1 := newAccount(1)
	account2 := newAccount(1, 0, 1)

	var gas uint64 = 100000

	//Shift right 0
	bytecode := MustSplice(PUSH1, 0x01, PUSH1, 0x00, SHR, return1())
	output, err := ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value := []uint8([]byte{0x1})
	expected := LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift right 0
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x00, SHR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Shift right 1
	bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0x01, SHR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift right 1
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0x01, SHR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x40})
	expected = RightPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift right 1
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SHR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected = []uint8([]byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Shift right 255
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0xFF, SHR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x1})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift right 255
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SHR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x1})
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
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x00})
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
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x00})
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
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

}

func TestSAR(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)
	account1 := newAccount(1)
	account2 := newAccount(1, 0, 1)

	var gas uint64 = 100000

	//Shift arith right 0
	bytecode := MustSplice(PUSH1, 0x01, PUSH1, 0x00, SAR, return1())
	output, err := ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value := []uint8([]byte{0x1})
	expected := LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative arith shift right 0
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x00, SAR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Shift arith right 1
	bytecode = MustSplice(PUSH1, 0x01, PUSH1, 0x01, SAR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x00})
	expected = LeftPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift arith right 1
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0x01, SAR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0xc0})
	expected = RightPadBytes(value, 32)
	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift arith right 1
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0x01, SAR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Shift arith right 255
	bytecode = MustSplice(PUSH32, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, PUSH1, 0xFF, SAR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift arith right 255
	bytecode = MustSplice(PUSH32, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SAR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

	//Alternative shift arith right 255
	bytecode = MustSplice(PUSH32, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, PUSH1, 0xFF, SAR, return1())
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
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
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
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
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	value = []uint8([]byte{0x00})
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
	output, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected = []uint8([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})

	assert.Equal(t, expected, output)

	t.Logf("Result: %v == %v\n", output, expected)

	if err != nil {
		t.Fatal(err)
	}

}

//Test attempt to jump to bad destination (position 16)
func TestJumpErr(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)

	// Create accounts
	account1 := newAccount(1)
	account2 := newAccount(2)

	var gas uint64 = 100000

	bytecode := MustSplice(PUSH1, 0x10, JUMP)

	var err error
	ch := make(chan struct{})
	go func() {
		_, err = ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
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
}

// Tests the code for a subcurrency contract compiled by serpent
func TestSubcurrency(t *testing.T) {
	st := newAppState()
	cache := state.NewCache(st)
	// Create accounts
	account1 := newAccount(1, 2, 3)
	account2 := newAccount(3, 2, 1)
	cache.UpdateAccount(account1)
	cache.UpdateAccount(account2)
	cache.Sync(st)
	//st.accounts[account1.Address()] = account1
	//st.accounts[account2.Address()] = account2

	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)

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
		JUMPDEST, POP, JUMPDEST, PUSH1, 0x00, RETURN)

	data, _ := hex.DecodeString("693200CE0000000000000000000000004B4363CDE27C2EB05E66357DB05BC5C88F850C1A0000000000000000000000000000000000000000000000000000000000000005")
	output, err := ourVm.Call(cache, account1, account2, bytecode, data, 0, &gas)
	fmt.Printf("Output: %v Error: %v\n", output, err)
	if err != nil {
		t.Fatal(err)
	}
}

//This test case is taken from EIP-140 (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-140.md);
//it is meant to test the implementation of the REVERT opcode
func TestRevert(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)

	// Create accounts
	account1 := newAccount(1)
	account2 := newAccount(1, 0, 1)

	key, value := []byte{0x00}, []byte{0x00}
	cache.SetStorage(account1.Address(), LeftPadWord256(key), LeftPadWord256(value))

	var gas uint64 = 100000

	bytecode := MustSplice(PUSH13, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x65, 0x64, 0x20, 0x64, 0x61, 0x74, 0x61,
		PUSH1, 0x00, SSTORE, PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, REVERT)

	/*bytecode := MustSplice(PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, REVERT)*/

	output, cErr := ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	assert.Error(t, cErr, "Expected execution reverted error")

	storageVal, err := cache.GetStorage(account1.Address(), LeftPadWord256(key))
	assert.Equal(t, LeftPadWord256(value), storageVal)

	t.Logf("Output: %v Error: %v\n", output, err)

}

// Test sending tokens from a contract to another account
func TestSendCall(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)

	// Create accounts
	account1 := newAccount(1)
	account2 := newAccount(2)
	account3 := newAccount(3)

	cache.UpdateAccount(account1)
	cache.UpdateAccount(account2)
	cache.UpdateAccount(account3)
	// account1 will call account2 which will trigger CALL opcode to account3
	addr := account3.Address()
	contractCode := callContractCode(addr)

	//----------------------------------------------
	// account2 has insufficient balance, should fail
	_, err := runVMWaitError(cache, ourVm, account1, account2, addr, contractCode, 1000)
	assert.Error(t, err, "Expected insufficient balance error")

	//----------------------------------------------
	// give account2 sufficient balance, should pass
	account2 = newAccount(2)
	err = account2.AddToBalance(100000)
	require.NoError(t, err)
	_, err = runVMWaitError(cache, ourVm, account1, account2, addr, contractCode, 1000)
	assert.NoError(t, err, "Should have sufficient balance")

	//----------------------------------------------
	// insufficient gas, should fail
	account2 = newAccount(2)
	err = account2.AddToBalance(100000)
	require.NoError(t, err)
	_, err = runVMWaitError(cache, ourVm, account1, account2, addr, contractCode, 100)
	assert.NoError(t, err, "Expected insufficient gas error")
}

// This test was introduced to cover an issues exposed in our handling of the
// gas limit passed from caller to callee on various forms of CALL.
// The idea of this test is to implement a simple DelegateCall in EVM code
// We first run the DELEGATECALL with _just_ enough gas expecting a simple return,
// and then run it with 1 gas unit less, expecting a failure
func TestDelegateCallGas(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)

	inOff := 0
	inSize := 0 // no call data
	retOff := 0
	retSize := 32
	calleeReturnValue := int64(20)

	// DELEGATECALL(retSize, refOffset, inSize, inOffset, addr, gasLimit)
	// 6 pops
	delegateCallCost := GasStackOp * 6
	// 1 push
	gasCost := GasStackOp
	// 2 pops, 1 push
	subCost := GasStackOp * 3
	pushCost := GasStackOp

	costBetweenGasAndDelegateCall := gasCost + subCost + delegateCallCost + pushCost

	// Do a simple operation using 1 gas unit
	calleeAccount, calleeAddress := makeAccountWithCode(cache, "callee",
		MustSplice(PUSH1, calleeReturnValue, return1()))

	// Here we split up the caller code so we can make a DELEGATE call with
	// different amounts of gas. The value we sandwich in the middle is the amount
	// we subtract from the available gas (that the caller has available), so:
	// code := MustSplice(callerCodePrefix, <amount to subtract from GAS> , callerCodeSuffix)
	// gives us the code to make the call
	callerCodePrefix := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize,
		PUSH1, inOff, PUSH20, calleeAddress, PUSH1)
	callerCodeSuffix := MustSplice(GAS, SUB, DELEGATECALL, returnWord())

	// Perform a delegate call
	callerAccount, _ := makeAccountWithCode(cache, "caller",
		MustSplice(callerCodePrefix,
			// Give just enough gas to make the DELEGATECALL
			costBetweenGasAndDelegateCall,
			callerCodeSuffix))

	// Should pass
	output, err := runVMWaitError(cache, ourVm, callerAccount, calleeAccount, calleeAddress,
		callerAccount.Code(), 100)
	assert.NoError(t, err, "Should have sufficient funds for call")
	assert.Equal(t, Int64ToWord256(calleeReturnValue).Bytes(), output)

	callerAccount.SetCode(MustSplice(callerCodePrefix,
		// Shouldn't be enough gas to make call
		costBetweenGasAndDelegateCall-1,
		callerCodeSuffix))

	// Should fail
	_, err = runVMWaitError(cache, ourVm, callerAccount, calleeAccount, calleeAddress,
		callerAccount.Code(), 100)
	assert.Error(t, err, "Should have insufficient gas for call")
}

func TestMemoryBounds(t *testing.T) {
	cache := state.NewCache(newAppState())
	memoryProvider := func() Memory {
		return NewDynamicMemory(1024, 2048)
	}
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger, MemoryProvider(memoryProvider))
	caller, _ := makeAccountWithCode(cache, "caller", nil)
	callee, _ := makeAccountWithCode(cache, "callee", nil)
	gas := uint64(100000)
	// This attempts to store a value at the memory boundary and return it
	word := One256
	output, err := ourVm.call(cache, caller, callee,
		MustSplice(pushWord(word), storeAtEnd(), MLOAD, storeAtEnd(), returnAfterStore()),
		nil, 0, &gas)
	assert.NoError(t, err)
	assert.Equal(t, word.Bytes(), output)

	// Same with number
	word = Int64ToWord256(232234234432)
	output, err = ourVm.call(cache, caller, callee,
		MustSplice(pushWord(word), storeAtEnd(), MLOAD, storeAtEnd(), returnAfterStore()),
		nil, 0, &gas)
	assert.NoError(t, err)
	assert.Equal(t, word.Bytes(), output)

	// Now test a series of boundary stores
	code := pushWord(word)
	for i := 0; i < 10; i++ {
		code = MustSplice(code, storeAtEnd(), MLOAD)
	}
	output, err = ourVm.call(cache, caller, callee, MustSplice(code, storeAtEnd(), returnAfterStore()),
		nil, 0, &gas)
	assert.NoError(t, err)
	assert.Equal(t, word.Bytes(), output)

	// Same as above but we should breach the upper memory limit set in memoryProvider
	code = pushWord(word)
	for i := 0; i < 100; i++ {
		code = MustSplice(code, storeAtEnd(), MLOAD)
	}
	output, err = ourVm.call(cache, caller, callee, MustSplice(code, storeAtEnd(), returnAfterStore()),
		nil, 0, &gas)
	assert.Error(t, err, "Should hit memory out of bounds")
}

func TestMsgSender(t *testing.T) {
	st := newAppState()
	cache := state.NewCache(st)
	account1 := newAccount(1, 2, 3)
	account2 := newAccount(3, 2, 1)
	cache.UpdateAccount(account1)
	cache.UpdateAccount(account2)
	cache.Sync(st)
	//st.accounts[account1.Address()] = account1
	//st.accounts[account2.Address()] = account2

	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)

	var gas uint64 = 100000

	/*
			pragma solidity ^0.4.0;

			contract SimpleStorage {
		                function get() public constant returns (address) {
		        	        return msg.sender;
		    	        }
			}
	*/

	// This bytecode is compiled from Solidity contract above using remix.ethereum.org online compiler
	code, err := hex.DecodeString("6060604052341561000f57600080fd5b60ca8061001d6000396000f30060606040526004361060" +
		"3f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680636d4ce63c14604457" +
		"5b600080fd5b3415604e57600080fd5b60546096565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ff" +
		"ffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6000339050905600a165627a" +
		"7a72305820b9ebf49535372094ae88f56d9ad18f2a79c146c8f56e7ef33b9402924045071e0029")
	require.NoError(t, err)

	// Run the contract initialisation code to obtain the contract code that would be mounted at account2
	contractCode, err := ourVm.Call(cache, account1, account2, code, code, 0, &gas)
	require.NoError(t, err)

	// Not needed for this test (since contract code is passed as argument to vm), but this is what an execution
	// framework must do
	account2.SetCode(contractCode)

	// Input is the function hash of `get()`
	input, err := hex.DecodeString("6d4ce63c")

	output, err := ourVm.Call(cache, account1, account2, contractCode, input, 0, &gas)
	require.NoError(t, err)

	assert.Equal(t, account1.Address().Word256().Bytes(), output)

}

func TestInvalid(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)

	// Create accounts
	account1 := newAccount(1)
	account2 := newAccount(1, 0, 1)

	var gas uint64 = 100000

	bytecode := MustSplice(PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61,
		0x67, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, INVALID)

	output, err := ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)
	expected := errors.ErrorCodeExecutionAborted.Error()
	assert.EqualError(t, err, expected)
	t.Logf("Output: %v Error: %v\n", output, err)

}

func TestReturnDataSize(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)

	accountName := "account2addresstests"

	callcode := MustSplice(PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61,
		0x67, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, RETURN)

	// Create accounts
	account1 := newAccount(1)
	account2, _ := makeAccountWithCode(cache, accountName, callcode)
	cache.UpdateAccount(account2)

	var gas uint64 = 100000

	gas1, gas2 := byte(0x1), byte(0x1)
	value := byte(0x69)
	inOff, inSize := byte(0x0), byte(0x0) // no call data
	retOff, retSize := byte(0x0), byte(0x0E)

	bytecode := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1, inOff, PUSH1, value, PUSH20,
		0x61, 0x63, 0x63, 0x6F, 0x75, 0x6E, 0x74, 0x32, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x74, 0x65,
		0x73, 0x74, 0x73, PUSH2, gas1, gas2, CALL, RETURNDATASIZE, PUSH1, 0x00, MSTORE, PUSH1, 0x20, PUSH1, 0x00, RETURN)

	expected := LeftPadBytes([]byte{0x0E}, 32)

	output, err := ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)

	assert.Equal(t, expected, output)

	t.Logf("Output: %v Error: %v\n", output, err)

	if err != nil {
		t.Fatal(err)
	}
}

func TestReturnDataCopy(t *testing.T) {
	cache := state.NewCache(newAppState())
	ourVm := NewVM(newParams(), crypto.ZeroAddress, nil, logger)

	accountName := "account2addresstests"

	callcode := MustSplice(PUSH32, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61,
		0x67, 0x65, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, PUSH1, 0x00, MSTORE, PUSH1, 0x0E, PUSH1, 0x00, RETURN)

	// Create accounts
	account1 := newAccount(1)
	account2, _ := makeAccountWithCode(cache, accountName, callcode)
	cache.UpdateAccount(account2)

	var gas uint64 = 100000

	gas1, gas2 := byte(0x1), byte(0x1)
	value := byte(0x69)
	inOff, inSize := byte(0x0), byte(0x0) // no call data
	retOff, retSize := byte(0x0), byte(0x0E)

	bytecode := MustSplice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1, inOff, PUSH1, value, PUSH20,
		0x61, 0x63, 0x63, 0x6F, 0x75, 0x6E, 0x74, 0x32, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x74, 0x65,
		0x73, 0x74, 0x73, PUSH2, gas1, gas2, CALL, RETURNDATASIZE, PUSH1, 0x00, PUSH1, 0x00, RETURNDATACOPY,
		RETURNDATASIZE, PUSH1, 0x00, RETURN)

	expected := []byte{0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x20, 0x6D, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65}

	output, err := ourVm.Call(cache, account1, account2, bytecode, []byte{}, 0, &gas)

	assert.Equal(t, expected, output)

	t.Logf("Output: %v Error: %v\n", output, err)

	if err != nil {
		t.Fatal(err)
	}
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

func makeAccountWithCode(accountUpdater state.AccountUpdater, name string,
	code []byte) (*acm.MutableAccount, crypto.Address) {
	address, _ := crypto.AddressFromBytes([]byte(name))
	account := acm.ConcreteAccount{
		Address:  address,
		Balance:  9999999,
		Code:     code,
		Sequence: 0,
	}.MutableAccount()
	accountUpdater.UpdateAccount(account)
	return account, account.Address()
}

// Subscribes to an AccCall, runs the vm, returns the output any direct exception
// and then waits for any exceptions transmitted by Data in the AccCall
// event (in the case of no direct error from call we will block waiting for
// at least 1 AccCall event)
func runVMWaitError(vmCache state.Cache, ourVm *VM, caller, callee *acm.MutableAccount, subscribeAddr crypto.Address,
	contractCode []byte, gas uint64) ([]byte, error) {
	eventCh := make(chan *evm_events.EventDataCall)
	output, err := runVM(eventCh, vmCache, ourVm, caller, callee, subscribeAddr, contractCode, gas)
	if err != nil {
		return output, err
	}
	select {
	case eventDataCall := <-eventCh:
		if eventDataCall.Exception != nil {
			return output, eventDataCall.Exception
		}
		return output, nil
	}
}

// Subscribes to an AccCall, runs the vm, returns the output and any direct
// exception
func runVM(eventCh chan<- *evm_events.EventDataCall, vmCache state.Cache, ourVm *VM, caller, callee *acm.MutableAccount,
	subscribeAddr crypto.Address, contractCode []byte, gas uint64) ([]byte, error) {

	// we need to catch the event from the CALL to check for exceptions
	emitter := event.NewEmitter(logging.NewNoopLogger())
	fmt.Printf("subscribe to %s\n", subscribeAddr)

	err := evm_events.SubscribeAccountCall(context.Background(), emitter, "test", subscribeAddr,
		nil, -1, eventCh)
	if err != nil {
		return nil, err
	}
	evc := event.NewEventCache(emitter)
	ourVm.SetPublisher(evc)
	start := time.Now()
	output, err := ourVm.Call(vmCache, caller, callee, contractCode, []byte{}, 0, &gas)
	fmt.Printf("Output: %v Error: %v\n", output, err)
	fmt.Println("Call took:", time.Since(start))
	evc.Flush()
	return output, err
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

func pushInt64(i int64) []byte {
	return pushWord(Int64ToWord256(i))
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

func TestPushWord(t *testing.T) {
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
}

// Kind of indirect test of Splice, but here to avoid import cycles
func TestBytecode(t *testing.T) {
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

	contractAccount := &acm.ConcreteAccount{Address: crypto.AddressFromWord256(Int64ToWord256(102))}
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
}

func TestConcat(t *testing.T) {
	assert.Equal(t,
		[]byte{0x01, 0x02, 0x03, 0x04},
		Concat([]byte{0x01, 0x02}, []byte{0x03, 0x04}))
}

func TestSubslice(t *testing.T) {
	const size = 10
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = byte(i)
	}
	for n := int64(0); n < size; n++ {
		data = data[:n]
		for offset := int64(-size); offset < size; offset++ {
			for length := int64(-size); length < size; length++ {
				_, ok := subslice(data, offset, length)
				if offset < 0 || length < 0 || n < offset {
					assert.False(t, ok)
				} else {
					assert.True(t, ok)
				}
			}
		}
	}
}

func TestHasPermission(t *testing.T) {
	st := newAppState()
	acc := acm.ConcreteAccount{
		Permissions: ptypes.AccountPermissions{
			Base: BasePermissionsFromStrings(t,
				"00100001000111",
				"11011110111000"),
		},
	}.Account()
	// Ensure we are falling through to global permissions on those bits not set
	assert.True(t, HasPermission(st, acc, PermFlagFromString(t, "100001000110")))
}

func BasePermissionsFromStrings(t *testing.T, perms, setBit string) ptypes.BasePermissions {
	return ptypes.BasePermissions{
		Perms:  PermFlagFromString(t, perms),
		SetBit: PermFlagFromString(t, setBit),
	}
}

func PermFlagFromString(t *testing.T, binaryString string) ptypes.PermFlag {
	permFlag, err := strconv.ParseUint(binaryString, 2, 64)
	require.NoError(t, err)
	return ptypes.PermFlag(permFlag)
}
