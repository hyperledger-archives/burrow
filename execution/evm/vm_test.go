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
	"encoding/hex"
	"fmt"
	//"strings"
	"testing"
	"time"

	acm "github.com/hyperledger/burrow/account"

	"errors"

	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/event"
	exe_events "github.com/hyperledger/burrow/execution/events"
	. "github.com/hyperledger/burrow/execution/evm/asm"
	. "github.com/hyperledger/burrow/execution/evm/asm/bc"
	evm_events "github.com/hyperledger/burrow/execution/evm/events"
	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var logger, _ = lifecycle.NewStdErrLogger()

func newAppState() *FakeAppState {
	fas := &FakeAppState{
		accounts: make(map[acm.Address]acm.Account),
		storage:  make(map[string]Word256),
	}
	// For default permissions
	fas.accounts[permission.GlobalPermissionsAddress] = acm.ConcreteAccount{
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

func newAccount(address ...byte) acm.MutableAccount {
	return acm.ConcreteAccount{
		Address: acm.AddressFromWord256(RightPadWord256(address)),
	}.MutableAccount()
}

// Runs a basic loop
func TestVM(t *testing.T) {
	ourVm := NewVM(newAppState(), DefaultDynamicMemoryProvider, newParams(), acm.ZeroAddress, nil, logger)

	// Create accounts
	account1 := newAccount(1)
	account2 := newAccount(1, 0, 1)

	var gas uint64 = 100000

	//Array defining how many times loop will run
	N := []byte{0x0f, 0x0f}

	// Loop initialization
	code := []byte{
		byte(PUSH1), 0x00, byte(PUSH1), 0x20, byte(MSTORE), byte(JUMPDEST),
		byte(0x60 + len(N) - 1),
	}

	code = append(code, N...)

	code = append(code, []byte{
		byte(PUSH1), 0x20, byte(MLOAD), byte(SLT), byte(ISZERO), byte(PUSH1),
		byte(0x1b + len(N)), byte(JUMPI), byte(PUSH1), 0x01, byte(PUSH1), 0x20,
		byte(MLOAD), byte(ADD), byte(PUSH1), 0x20, byte(MSTORE), byte(PUSH1),
		0x05, byte(JUMP), byte(JUMPDEST),
	}...)

	start := time.Now()
	output, err := ourVm.Call(account1, account2, code, []byte{}, 0, &gas)
	fmt.Printf("Output: %v Error: %v\n", output, err)
	fmt.Println("Call took:", time.Since(start))
	if err != nil {
		t.Fatal(err)
	}
}

//Test attempt to jump to bad destination (position 16)
func TestJumpErr(t *testing.T) {
	ourVm := NewVM(newAppState(), DefaultDynamicMemoryProvider, newParams(), acm.ZeroAddress, nil, logger)

	// Create accounts
	account1 := newAccount(1)
	account2 := newAccount(2)

	var gas uint64 = 100000

	//Set jump destination to 16
	code := []byte{byte(PUSH1), 0x10, byte(JUMP)}

	var err error
	ch := make(chan struct{})
	go func() {
		_, err = ourVm.Call(account1, account2, code, []byte{}, 0, &gas)
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

	// Create accounts
	account1 := newAccount(1, 2, 3)
	account2 := newAccount(3, 2, 1)
	st.accounts[account1.Address()] = account1
	st.accounts[account2.Address()] = account2

	ourVm := NewVM(st, DefaultDynamicMemoryProvider, newParams(), acm.ZeroAddress, nil, logger)

	var gas uint64 = 1000

	code := []byte{byte(PUSH3), 0x0f, 0x42, 0x40, byte(CALLER), byte(SSTORE)}
	code = append(code, []byte{
		byte(PUSH29), 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}...)
	code = append(code, []byte{
		byte(PUSH1), 0x00, byte(CALLDATALOAD), byte(DIV), byte(PUSH4), 0x15, 0xcf,
		0x26, 0x84, byte(DUP2), byte(EQ), byte(ISZERO), byte(PUSH2), 0x00, 0x46, byte(JUMPI),
	}...)
	code = append(code, []byte{
		byte(PUSH1), 0x04, byte(CALLDATALOAD), byte(PUSH1), 0x40, byte(MSTORE),
		byte(PUSH1), 0x40, byte(MLOAD), byte(SLOAD), byte(PUSH1), 0x60, byte(MSTORE),
		byte(PUSH1), 0x20, byte(PUSH1), 0x60, byte(RETURN), byte(JUMPDEST),
		byte(PUSH4), 0x69, 0x32, 0x00, 0xce, byte(DUP2), byte(EQ), byte(ISZERO),
		byte(PUSH2), 0x00, 0x87, byte(JUMPI),
	}...)
	code = append(code, []byte{
		byte(PUSH1), 0x04, byte(CALLDATALOAD), byte(PUSH1), 0x80, byte(MSTORE),
		byte(PUSH1), 0x24, byte(CALLDATALOAD), byte(PUSH1), 0xa0, byte(MSTORE),
		byte(CALLER), byte(SLOAD), byte(PUSH1), 0xc0, byte(MSTORE), byte(CALLER),
		byte(PUSH1), 0xe0, byte(MSTORE), byte(PUSH1), 0xa0, byte(MLOAD),
		byte(PUSH1), 0xc0, byte(MLOAD), byte(SLT), byte(ISZERO), byte(ISZERO),
		byte(PUSH2), 0x00, 0x86, byte(JUMPI),
	}...)
	code = append(code, []byte{
		byte(PUSH1), 0xa0, byte(MLOAD), byte(PUSH1), 0xc0, byte(MLOAD), byte(SUB),
		byte(PUSH1), 0xe0, byte(MLOAD), byte(SSTORE), byte(PUSH1), 0xa0, byte(MLOAD),
		byte(PUSH1), 0x80, byte(MLOAD), byte(SLOAD), byte(ADD), byte(PUSH1), 0x80,
		byte(MLOAD), byte(SSTORE), byte(JUMPDEST), byte(JUMPDEST), byte(POP),
		byte(JUMPDEST), byte(PUSH1), 0x00, byte(RETURN),
	}...)

	for _, element := range code {
		fmt.Printf("Code: %#x\n", element)
	}

	data, _ := hex.DecodeString("693200CE0000000000000000000000004B4363CDE27C2EB05E66357DB05BC5C88F850C1A0000000000000000000000000000000000000000000000000000000000000005")
	output, err := ourVm.Call(account1, account2, code, data, 0, &gas)
	fmt.Printf("Output: %v Error: %v\n", output, err)
	if err != nil {
		t.Fatal(err)
	}
}

// Test sending tokens from a contract to another account
func TestSendCall(t *testing.T) {
	fakeAppState := newAppState()
	ourVm := NewVM(fakeAppState, DefaultDynamicMemoryProvider, newParams(), acm.ZeroAddress, nil, logger)

	// Create accounts
	account1 := newAccount(1)
	account2 := newAccount(2)
	account3 := newAccount(3)

	// account1 will call account2 which will trigger CALL opcode to account3
	addr := account3.Address()
	contractCode := callContractCode(addr)

	//----------------------------------------------
	// account2 has insufficient balance, should fail
	_, err := runVMWaitError(ourVm, account1, account2, addr, contractCode, 1000)
	assert.Error(t, err, "Expected insufficient balance error")

	//----------------------------------------------
	// give account2 sufficient balance, should pass
	account2, err = newAccount(2).AddToBalance(100000)
	require.NoError(t, err)
	_, err = runVMWaitError(ourVm, account1, account2, addr, contractCode, 1000)
	assert.NoError(t, err, "Should have sufficient balance")

	//----------------------------------------------
	// insufficient gas, should fail
	account2, err = newAccount(2).AddToBalance(100000)
	require.NoError(t, err)
	_, err = runVMWaitError(ourVm, account1, account2, addr, contractCode, 100)
	assert.Error(t, err, "Expected insufficient gas error")
}

// This test was introduced to cover an issues exposed in our handling of the
// gas limit passed from caller to callee on various forms of CALL.
// The idea of this test is to implement a simple DelegateCall in EVM code
// We first run the DELEGATECALL with _just_ enough gas expecting a simple return,
// and then run it with 1 gas unit less, expecting a failure
func TestDelegateCallGas(t *testing.T) {
	state := newAppState()
	ourVm := NewVM(state, DefaultDynamicMemoryProvider, newParams(), acm.ZeroAddress, nil, logger)

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
	calleeAccount, calleeAddress := makeAccountWithCode(state, "callee",
		Splice(PUSH1, calleeReturnValue, return1()))

	// Here we split up the caller code so we can make a DELEGATE call with
	// different amounts of gas. The value we sandwich in the middle is the amount
	// we subtract from the available gas (that the caller has available), so:
	// code := Splice(callerCodePrefix, <amount to subtract from GAS> , callerCodeSuffix)
	// gives us the code to make the call
	callerCodePrefix := Splice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize,
		PUSH1, inOff, PUSH20, calleeAddress, PUSH1)
	callerCodeSuffix := Splice(GAS, SUB, DELEGATECALL, returnWord())

	// Perform a delegate call
	callerAccount, _ := makeAccountWithCode(state, "caller",
		Splice(callerCodePrefix,
			// Give just enough gas to make the DELEGATECALL
			costBetweenGasAndDelegateCall,
			callerCodeSuffix))

	// Should pass
	output, err := runVMWaitError(ourVm, callerAccount, calleeAccount, calleeAddress,
		callerAccount.Code(), 100)
	assert.NoError(t, err, "Should have sufficient funds for call")
	assert.Equal(t, Int64ToWord256(calleeReturnValue).Bytes(), output)

	callerAccount.SetCode(Splice(callerCodePrefix,
		// Shouldn't be enough gas to make call
		costBetweenGasAndDelegateCall-1,
		callerCodeSuffix))

	// Should fail
	_, err = runVMWaitError(ourVm, callerAccount, calleeAccount, calleeAddress,
		callerAccount.Code(), 100)
	assert.Error(t, err, "Should have insufficient funds for call")
}

func TestMemoryBounds(t *testing.T) {
	state := newAppState()
	memoryProvider := func() Memory {
		return NewDynamicMemory(1024, 2048)
	}
	ourVm := NewVM(state, memoryProvider, newParams(), acm.ZeroAddress, nil, logger)
	caller, _ := makeAccountWithCode(state, "caller", nil)
	callee, _ := makeAccountWithCode(state, "callee", nil)
	gas := uint64(100000)
	// This attempts to store a value at the memory boundary and return it
	word := One256
	output, err := ourVm.call(caller, callee,
		Splice(pushWord(word), storeAtEnd(), MLOAD, storeAtEnd(), returnAfterStore()),
		nil, 0, &gas)
	assert.NoError(t, err)
	assert.Equal(t, word.Bytes(), output)

	// Same with number
	word = Int64ToWord256(232234234432)
	output, err = ourVm.call(caller, callee,
		Splice(pushWord(word), storeAtEnd(), MLOAD, storeAtEnd(), returnAfterStore()),
		nil, 0, &gas)
	assert.NoError(t, err)
	assert.Equal(t, word.Bytes(), output)

	// Now test a series of boundary stores
	code := pushWord(word)
	for i := 0; i < 10; i++ {
		code = Splice(code, storeAtEnd(), MLOAD)
	}
	output, err = ourVm.call(caller, callee, Splice(code, storeAtEnd(), returnAfterStore()),
		nil, 0, &gas)
	assert.NoError(t, err)
	assert.Equal(t, word.Bytes(), output)

	// Same as above but we should breach the upper memory limit set in memoryProvider
	code = pushWord(word)
	for i := 0; i < 100; i++ {
		code = Splice(code, storeAtEnd(), MLOAD)
	}
	output, err = ourVm.call(caller, callee, Splice(code, storeAtEnd(), returnAfterStore()),
		nil, 0, &gas)
	assert.Error(t, err, "Should hit memory out of bounds")
}

// These code segment helpers exercise the MSTORE MLOAD MSTORE cycle to test
// both of the memory operations. Each MSTORE is done on the memory boundary
// (at MSIZE) which Solidity uses to find guaranteed unallocated memory.

// storeAtEnd expects the value to be stored to be on top of the stack, it then
// stores that value at the current memory boundary
func storeAtEnd() []byte {
	// Pull in MSIZE (to carry forward to MLOAD), swap in value to store, store it at MSIZE
	return Splice(MSIZE, SWAP1, DUP2, MSTORE)
}

func returnAfterStore() []byte {
	return Splice(PUSH1, 32, DUP2, RETURN)
}

// Store the top element of the stack (which is a 32-byte word) in memory
// and return it. Useful for a simple return value.
func return1() []byte {
	return Splice(PUSH1, 0, MSTORE, returnWord())
}

func returnWord() []byte {
	// PUSH1 => return size, PUSH1 => return offset, RETURN
	return Splice(PUSH1, 32, PUSH1, 0, RETURN)
}

func makeAccountWithCode(state acm.Updater, name string,
	code []byte) (acm.MutableAccount, acm.Address) {
	address, _ := acm.AddressFromBytes([]byte(name))
	account := acm.ConcreteAccount{
		Address:  address,
		Balance:  9999999,
		Code:     code,
		Sequence: 0,
	}.MutableAccount()
	state.UpdateAccount(account)
	return account, account.Address()
}

// Subscribes to an AccCall, runs the vm, returns the output any direct exception
// and then waits for any exceptions transmitted by Data in the AccCall
// event (in the case of no direct error from call we will block waiting for
// at least 1 AccCall event)
func runVMWaitError(ourVm *VM, caller, callee acm.MutableAccount, subscribeAddr acm.Address,
	contractCode []byte, gas uint64) (output []byte, err error) {
	eventCh := make(chan event.EventData)
	output, err = runVM(eventCh, ourVm, caller, callee, subscribeAddr,
		contractCode, gas)
	if err != nil {
		return
	}
	msg := <-eventCh
	var errString string
	switch ev := msg.Unwrap().(type) {
	case exe_events.EventDataTx:
		errString = ev.Exception
	case evm_events.EventDataCall:
		errString = ev.Exception
	}

	if errString != "" {
		err = errors.New(errString)
	}
	return
}

// Subscribes to an AccCall, runs the vm, returns the output and any direct
// exception
func runVM(eventCh chan event.EventData, ourVm *VM, caller, callee acm.MutableAccount,
	subscribeAddr acm.Address, contractCode []byte, gas uint64) ([]byte, error) {

	// we need to catch the event from the CALL to check for exceptions
	evsw := event.NewEmitter(loggers.NewNoopInfoTraceLogger())
	fmt.Printf("subscribe to %s\n", subscribeAddr)
	evsw.Subscribe("test", evm_events.EventStringAccCall(subscribeAddr),
		func(msg event.AnyEventData) {
			eventCh <- *msg.BurrowEventData
		})
	evc := event.NewEventCache(evsw)
	ourVm.SetFireable(evc)
	start := time.Now()
	output, err := ourVm.Call(caller, callee, contractCode, []byte{}, 0, &gas)
	fmt.Printf("Output: %v Error: %v\n", output, err)
	fmt.Println("Call took:", time.Since(start))
	go func() { evc.Flush() }()
	return output, err
}

// this is code to call another contract (hardcoded as addr)
func callContractCode(addr acm.Address) []byte {
	gas1, gas2 := byte(0x1), byte(0x1)
	value := byte(0x69)
	inOff, inSize := byte(0x0), byte(0x0) // no call data
	retOff, retSize := byte(0x0), byte(0x20)
	// this is the code we want to run (send funds to an account and return)
	return Splice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1,
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
			return Splice(byte(PUSH32)-leadingZeros, word[leadingZeros:])
		}
	}
	return Splice(PUSH1, 0)
}

func TestPushWord(t *testing.T) {
	word := Int64ToWord256(int64(2133213213))
	assert.Equal(t, Splice(PUSH4, 0x7F, 0x26, 0x40, 0x1D), pushWord(word))
	word[0] = 1
	assert.Equal(t, Splice(PUSH32,
		1, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0x7F, 0x26, 0x40, 0x1D), pushWord(word))
	assert.Equal(t, Splice(PUSH1, 0), pushWord(Word256{}))
	assert.Equal(t, Splice(PUSH1, 1), pushWord(Int64ToWord256(1)))
}

func TestBytecode(t *testing.T) {
	assert.Equal(t,
		Splice(1, 2, 3, 4, 5, 6),
		Splice(1, 2, 3, Splice(4, 5, 6)))
	assert.Equal(t,
		Splice(1, 2, 3, 4, 5, 6, 7, 8),
		Splice(1, 2, 3, Splice(4, Splice(5), 6), 7, 8))
	assert.Equal(t,
		Splice(PUSH1, 2),
		Splice(byte(PUSH1), 0x02))
	assert.Equal(t,
		[]byte{},
		Splice(Splice(Splice())))

	contractAccount := &acm.ConcreteAccount{Address: acm.AddressFromWord256(Int64ToWord256(102))}
	addr := contractAccount.Address
	gas1, gas2 := byte(0x1), byte(0x1)
	value := byte(0x69)
	inOff, inSize := byte(0x0), byte(0x0) // no call data
	retOff, retSize := byte(0x0), byte(0x20)
	contractCodeBytecode := Splice(PUSH1, retSize, PUSH1, retOff, PUSH1, inSize, PUSH1,
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
