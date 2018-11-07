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
	"bytes"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	. "github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/sha3"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
)

const (
	DataStackInitialCapacity = 1024
	callStackCapacity        = 100 // TODO ensure usage.
)

type Params struct {
	BlockHeight              uint64
	BlockHash                Word256
	BlockTime                int64
	GasLimit                 uint64
	CallStackMaxDepth        uint64
	DataStackInitialCapacity uint64
	DataStackMaxDepth        uint64
}

type VM struct {
	memoryProvider func() Memory
	params         Params
	origin         crypto.Address
	tx             *txs.Tx
	stackDepth     uint64
	logger         *logging.Logger
	returnData     []byte
	debugOpcodes   bool
	dumpTokens     bool
}

func NewVM(params Params, origin crypto.Address, tx *txs.Tx, logger *logging.Logger, options ...func(*VM)) *VM {
	vm := &VM{
		memoryProvider: DefaultDynamicMemoryProvider,
		params:         params,
		origin:         origin,
		stackDepth:     0,
		tx:             tx,
		logger:         logger.WithScope("NewVM"),
	}
	for _, option := range options {
		option(vm)
	}
	return vm
}

func (vm *VM) Debugf(format string, a ...interface{}) {
	// Uncomment for quick and dirty debug
	//fmt.Printf(format, a...)
	if vm.debugOpcodes {
		vm.logger.TraceMsg(fmt.Sprintf(format, a...), "tag", "DebugOpcodes")
	}
}

// CONTRACT: it is the duty of the contract writer to call known permissions
// we do not convey if a permission is not set
// (unlike in state/execution, where we guarantee HasPermission is called
// on known permissions and panics else)
// If the perm is not defined in the acc nor set by default in GlobalPermissions,
// this function returns false.
func HasPermission(st Interface, address crypto.Address, perm permission.PermFlag) bool {
	globalPerms := st.GetPermissions(acm.GlobalPermissionsAddress)
	accPerms := st.GetPermissions(address)
	perms := accPerms.Base.Compose(globalPerms.Base)
	value, err := perms.Get(perm)
	if err != nil {
		return false
	}
	return value
}

func EnsurePermission(st Interface, address crypto.Address, perm permission.PermFlag) errors.CodedError {
	if !HasPermission(st, address, perm) {
		return errors.PermissionDenied{
			Address: address,
			Perm:    perm,
		}
	}
	return nil
}

func (vm *VM) fireCallEvent(eventSink EventSink, callType exec.CallType, exception *errors.CodedError, output *[]byte,
	callerAddress, calleeAddress crypto.Address, input []byte, value uint64, gas *uint64, err *errors.CodedError) {
	// fire the post call event (including exception if applicable)
	eventErr := eventSink.Call(&exec.CallEvent{
		CallType: callType,
		CallData: &exec.CallData{
			Caller: callerAddress,
			Callee: calleeAddress,
			Data:   input,
			Value:  value,
			Gas:    *gas,
		},
		Origin:     vm.origin,
		StackDepth: vm.stackDepth,
		Return:     *output,
	}, errors.AsException(*exception))
	if eventErr != nil {
		*err = firstErr(*err, eventErr)
	}
}

// CONTRACT state is aware of caller and callee, so we can just mutate them.
// CONTRACT code and input are not mutated.
// CONTRACT returned 'ret' is a new compact slice.
// value: To be transferred from caller to callee. Refunded upon errors.CodedError.
// gas:   Available gas. No refunds for gas.
// code: May be nil, since the CALL opcode may be used to send value from contracts to accounts
func (vm *VM) Call(callState Interface, eventSink EventSink, caller, callee crypto.Address, code,
	input []byte, value uint64, gas *uint64) (output []byte, err errors.CodedError) {

	// Always return output - we may have a reverted exception for which the return is meaningful
	output, err = vm.call(callState, eventSink, caller, callee, code, input, value, gas, exec.CallTypeCall)
	if err == nil {
		err = callState.Error()
	}
	return
}

func (vm *VM) call(callState Interface, eventSink EventSink, caller, callee crypto.Address, code,
	input []byte, value uint64, gas *uint64, callType exec.CallType) (output []byte, err errors.CodedError) {

	exception := new(errors.CodedError)
	// fire the post call event (including exception if applicable)
	defer vm.fireCallEvent(eventSink, callType, exception, &output, caller, callee, input,
		value, gas, &err)

	if err = transfer(callState, caller, callee, value); err != nil {
		*exception = err
		return
	}

	if err = vm.ensureStackDepth(); err != nil {
		*exception = err
		return
	}

	if len(code) > 0 {
		vm.stackDepth += 1
		output, err = vm.execute(callState, eventSink, caller, callee, code, input, value, gas)
		vm.stackDepth -= 1
		if err != nil {
			*exception = err
			transferErr := transfer(callState, callee, caller, value)
			if transferErr != nil {
				err = errors.Wrap(transferErr,
					fmt.Sprintf("error refunding value %v %s (callee) -> %s (caller)", value, callee, caller))
			}
		}
	}

	return
}

// DelegateCall is executed by the DELEGATECALL opcode, introduced as off Ethereum Homestead.
// The intent of delegate call is to run the code of the callee in the storage context of the caller;
// while preserving the original caller to the previous callee.
// Different to the normal CALL or CALLCODE, the value does not need to be transferred to the callee.
func (vm *VM) delegateCall(callState Interface, eventSink EventSink, caller, callee crypto.Address,
	code, input []byte, value uint64, gas *uint64,
	callType exec.CallType) (output []byte, err errors.CodedError) {

	exception := new(errors.CodedError)

	// &err allows fireCallEvent to push back an error to this function's return
	defer vm.fireCallEvent(eventSink, callType, exception, &output, caller, callee, input, value, gas, &err)

	// DelegateCall does not transfer the value to the callee.

	if err = vm.ensureStackDepth(); err != nil {
		*exception = err
		return
	}

	if len(code) > 0 {
		vm.stackDepth += 1
		output, err = vm.execute(callState, eventSink, caller, callee, code, input, value, gas)
		vm.stackDepth -= 1
		if err != nil {
			*exception = err
		}
	}

	return
}

// Try to deduct gasToUse from gasLeft.  If ok return false, otherwise
// set err and return true.
func useGasNegative(gasLeft *uint64, gasToUse uint64, err *errors.CodedError) bool {
	if *gasLeft >= gasToUse {
		*gasLeft -= gasToUse
		return false
	} else if *err == nil {
		*err = errors.ErrorCodeInsufficientGas
	}
	return true
}

// Executes the EVM code passed in the appropriate context
func (vm *VM) execute(callState Interface, eventSink EventSink, caller, callee crypto.Address,
	code, input []byte, value uint64, gas *uint64) (output []byte, err errors.CodedError) {
	vm.Debugf("(%d) (%s) %s (code=%d) gas: %v (d) %X\n", vm.stackDepth, caller, callee, len(code), *gas, input)

	logger := vm.logger.With("tx_hash", vm.tx.Hash())

	if vm.dumpTokens {
		dumpTokens(vm.tx.Hash(), caller, callee, code)
	}

	var (
		pc     int64 = 0
		stack        = NewStack(vm.params.DataStackInitialCapacity, vm.params.DataStackMaxDepth, gas, &err)
		memory       = vm.memoryProvider()
	)

	for {
		// Check for any error accrued to state
		if callState.Error() != nil {
			return nil, firstErr(err, callState.Error())
		}
		// Use BaseOp gas.
		if useGasNegative(gas, GasBaseOp, &err) {
			return nil, err
		}

		var op = codeGetOp(code, pc)
		vm.Debugf("(pc) %-3d (op) %-14s (st) %-4d ", pc, op.String(), stack.Len())

		switch op {

		case ADD: // 0x01
			x, y := stack.PopBigInt(), stack.PopBigInt()
			sum := new(big.Int).Add(x, y)
			res := stack.PushBigInt(sum)
			vm.Debugf(" %v + %v = %v (%X)\n", x, y, sum, res)

		case MUL: // 0x02
			x, y := stack.PopBigInt(), stack.PopBigInt()
			prod := new(big.Int).Mul(x, y)
			res := stack.PushBigInt(prod)
			vm.Debugf(" %v * %v = %v (%X)\n", x, y, prod, res)

		case SUB: // 0x03
			x, y := stack.PopBigInt(), stack.PopBigInt()
			diff := new(big.Int).Sub(x, y)
			res := stack.PushBigInt(diff)
			vm.Debugf(" %v - %v = %v (%X)\n", x, y, diff, res)

		case DIV: // 0x04
			x, y := stack.PopBigInt(), stack.PopBigInt()
			if y.Sign() == 0 {
				stack.Push(Zero256)
				vm.Debugf(" %x / %x = %v\n", x, y, 0)
			} else {
				div := new(big.Int).Div(x, y)
				res := stack.PushBigInt(div)
				vm.Debugf(" %v / %v = %v (%X)\n", x, y, div, res)
			}

		case SDIV: // 0x05
			x, y := stack.PopBigIntSigned(), stack.PopBigIntSigned()
			if y.Sign() == 0 {
				stack.Push(Zero256)
				vm.Debugf(" %x / %x = %v\n", x, y, 0)
			} else {
				div := new(big.Int).Div(x, y)
				res := stack.PushBigInt(div)
				vm.Debugf(" %v / %v = %v (%X)\n", x, y, div, res)
			}

		case MOD: // 0x06
			x, y := stack.PopBigInt(), stack.PopBigInt()
			if y.Sign() == 0 {
				stack.Push(Zero256)
				vm.Debugf(" %v %% %v = %v\n", x, y, 0)
			} else {
				mod := new(big.Int).Mod(x, y)
				res := stack.PushBigInt(mod)
				vm.Debugf(" %v %% %v = %v (%X)\n", x, y, mod, res)
			}

		case SMOD: // 0x07
			x, y := stack.PopBigIntSigned(), stack.PopBigIntSigned()
			if y.Sign() == 0 {
				stack.Push(Zero256)
				vm.Debugf(" %v %% %v = %v\n", x, y, 0)
			} else {
				mod := new(big.Int).Mod(x, y)
				res := stack.PushBigInt(mod)
				vm.Debugf(" %v %% %v = %v (%X)\n", x, y, mod, res)
			}

		case ADDMOD: // 0x08
			x, y, z := stack.PopBigInt(), stack.PopBigInt(), stack.PopBigInt()
			if z.Sign() == 0 {
				stack.Push(Zero256)
				vm.Debugf(" %v %% %v = %v\n", x, y, 0)
			} else {
				add := new(big.Int).Add(x, y)
				mod := add.Mod(add, z)
				res := stack.PushBigInt(mod)
				vm.Debugf(" %v + %v %% %v = %v (%X)\n", x, y, z, mod, res)
			}

		case MULMOD: // 0x09
			x, y, z := stack.PopBigInt(), stack.PopBigInt(), stack.PopBigInt()
			if z.Sign() == 0 {
				stack.Push(Zero256)
				vm.Debugf(" %v %% %v = %v\n", x, y, 0)
			} else {
				mul := new(big.Int).Mul(x, y)
				mod := mul.Mod(mul, z)
				res := stack.PushBigInt(mod)
				vm.Debugf(" %v * %v %% %v = %v (%X)\n", x, y, z, mod, res)
			}

		case EXP: // 0x0A
			x, y := stack.PopBigInt(), stack.PopBigInt()
			pow := new(big.Int).Exp(x, y, nil)
			res := stack.PushBigInt(pow)
			vm.Debugf(" %v ** %v = %v (%X)\n", x, y, pow, res)

		case SIGNEXTEND: // 0x0B
			back, popErr := stack.PopU64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			if back < Word256Length-1 {
				stack.PushBigInt(SignExtend(back, stack.PopBigInt()))
			}

		case LT: // 0x10
			x, y := stack.PopBigInt(), stack.PopBigInt()
			if x.Cmp(y) < 0 {
				stack.Push(One256)
				vm.Debugf(" %v < %v = %v\n", x, y, 1)
			} else {
				stack.Push(Zero256)
				vm.Debugf(" %v < %v = %v\n", x, y, 0)
			}

		case GT: // 0x11
			x, y := stack.PopBigInt(), stack.PopBigInt()
			if x.Cmp(y) > 0 {
				stack.Push(One256)
				vm.Debugf(" %v > %v = %v\n", x, y, 1)
			} else {
				stack.Push(Zero256)
				vm.Debugf(" %v > %v = %v\n", x, y, 0)
			}

		case SLT: // 0x12
			x, y := stack.PopBigIntSigned(), stack.PopBigIntSigned()
			if x.Cmp(y) < 0 {
				stack.Push(One256)
				vm.Debugf(" %v < %v = %v\n", x, y, 1)
			} else {
				stack.Push(Zero256)
				vm.Debugf(" %v < %v = %v\n", x, y, 0)
			}

		case SGT: // 0x13
			x, y := stack.PopBigIntSigned(), stack.PopBigIntSigned()
			if x.Cmp(y) > 0 {
				stack.Push(One256)
				vm.Debugf(" %v > %v = %v\n", x, y, 1)
			} else {
				stack.Push(Zero256)
				vm.Debugf(" %v > %v = %v\n", x, y, 0)
			}

		case EQ: // 0x14
			x, y := stack.Pop(), stack.Pop()
			if bytes.Equal(x[:], y[:]) {
				stack.Push(One256)
				vm.Debugf(" %X == %X = %v\n", x, y, 1)
			} else {
				stack.Push(Zero256)
				vm.Debugf(" %X == %X = %v\n", x, y, 0)
			}

		case ISZERO: // 0x15
			x := stack.Pop()
			if x.IsZero() {
				stack.Push(One256)
				vm.Debugf(" %X == 0 = %v\n", x, 1)
			} else {
				stack.Push(Zero256)
				vm.Debugf(" %X == 0 = %v\n", x, 0)
			}

		case AND: // 0x16
			x, y := stack.Pop(), stack.Pop()
			z := [32]byte{}
			for i := 0; i < 32; i++ {
				z[i] = x[i] & y[i]
			}
			stack.Push(z)
			vm.Debugf(" %X & %X = %X\n", x, y, z)

		case OR: // 0x17
			x, y := stack.Pop(), stack.Pop()
			z := [32]byte{}
			for i := 0; i < 32; i++ {
				z[i] = x[i] | y[i]
			}
			stack.Push(z)
			vm.Debugf(" %X | %X = %X\n", x, y, z)

		case XOR: // 0x18
			x, y := stack.Pop(), stack.Pop()
			z := [32]byte{}
			for i := 0; i < 32; i++ {
				z[i] = x[i] ^ y[i]
			}
			stack.Push(z)
			vm.Debugf(" %X ^ %X = %X\n", x, y, z)

		case NOT: // 0x19
			x := stack.Pop()
			z := [32]byte{}
			for i := 0; i < 32; i++ {
				z[i] = ^x[i]
			}
			stack.Push(z)
			vm.Debugf(" !%X = %X\n", x, z)

		case BYTE: // 0x1A
			idx, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			val := stack.Pop()
			res := byte(0)
			if idx < 32 {
				res = val[idx]
			}
			stack.Push64(int64(res))
			vm.Debugf(" => 0x%X\n", res)

		case SHL: //0x1B
			shift, x := stack.PopBigInt(), stack.PopBigInt()

			if shift.Cmp(Big256) >= 0 {
				reset := big.NewInt(0)
				stack.PushBigInt(reset)
				vm.Debugf(" %v << %v = %v\n", x, shift, reset)
			} else {
				shiftedValue := x.Lsh(x, uint(shift.Uint64()))
				stack.PushBigInt(shiftedValue)
				vm.Debugf(" %v << %v = %v\n", x, shift, shiftedValue)
			}

		case SHR: //0x1C
			shift, x := stack.PopBigInt(), stack.PopBigInt()

			if shift.Cmp(Big256) >= 0 {
				reset := big.NewInt(0)
				stack.PushBigInt(reset)
				vm.Debugf(" %v << %v = %v\n", x, shift, reset)
			} else {
				shiftedValue := x.Rsh(x, uint(shift.Uint64()))
				stack.PushBigInt(shiftedValue)
				vm.Debugf(" %v << %v = %v\n", x, shift, shiftedValue)
			}

		case SAR: //0x1D
			shift, x := stack.PopBigInt(), stack.PopBigIntSigned()

			if shift.Cmp(Big256) >= 0 {
				reset := big.NewInt(0)
				if x.Sign() < 0 {
					reset.SetInt64(-1)
				}
				stack.PushBigInt(reset)
				vm.Debugf(" %v << %v = %v\n", x, shift, reset)
			} else {
				shiftedValue := x.Rsh(x, uint(shift.Uint64()))
				stack.PushBigInt(shiftedValue)
				vm.Debugf(" %v << %v = %v\n", x, shift, shiftedValue)
			}

		case SHA3: // 0x20
			if useGasNegative(gas, GasSha3, &err) {
				return nil, err
			}
			offset, size := stack.PopBigInt(), stack.PopBigInt()
			data, memErr := memory.Read(offset, size)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}
			data = sha3.Sha3(data)
			stack.PushBytes(data)
			vm.Debugf(" => (%v) %X\n", size, data)

		case ADDRESS: // 0x30
			stack.Push(callee.Word256())
			vm.Debugf(" => %X\n", callee)

		case BALANCE: // 0x31
			addr := stack.Pop()
			if useGasNegative(gas, GasGetAccount, &err) {
				return nil, err
			}
			balance := callState.GetBalance(crypto.AddressFromWord256(addr))
			stack.PushU64(balance)
			vm.Debugf(" => %v (%X)\n", balance, addr)

		case ORIGIN: // 0x32
			stack.Push(vm.origin.Word256())
			vm.Debugf(" => %X\n", vm.origin)

		case CALLER: // 0x33
			stack.Push(caller.Word256())
			vm.Debugf(" => %X\n", caller)

		case CALLVALUE: // 0x34
			stack.PushU64(value)
			vm.Debugf(" => %v\n", value)

		case CALLDATALOAD: // 0x35
			offset, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			data, ok := subslice(input, offset, 32)
			if !ok {
				return nil, firstErr(err, errors.ErrorCodeInputOutOfBounds)
			}
			res := LeftPadWord256(data)
			stack.Push(res)
			vm.Debugf(" => 0x%X\n", res)

		case CALLDATASIZE: // 0x36
			stack.Push64(int64(len(input)))
			vm.Debugf(" => %d\n", len(input))

		case CALLDATACOPY: // 0x37
			memOff := stack.PopBigInt()
			inputOff, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			length, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			data, ok := subslice(input, inputOff, length)
			if !ok {
				return nil, firstErr(err, errors.ErrorCodeInputOutOfBounds)
			}
			memErr := memory.Write(memOff, data)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}
			vm.Debugf(" => [%v, %v, %v] %X\n", memOff, inputOff, length, data)

		case CODESIZE: // 0x38
			l := int64(len(code))
			stack.Push64(l)
			vm.Debugf(" => %d\n", l)

		case CODECOPY: // 0x39
			memOff := stack.PopBigInt()
			codeOff, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			length, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			data, ok := subslice(code, codeOff, length)
			if !ok {
				return nil, firstErr(err, errors.ErrorCodeCodeOutOfBounds)
			}
			memErr := memory.Write(memOff, data)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}
			vm.Debugf(" => [%v, %v, %v] %X\n", memOff, codeOff, length, data)

		case GASPRICE_DEPRECATED: // 0x3A
			stack.Push(Zero256)
			vm.Debugf(" => %X (GASPRICE IS DEPRECATED)\n", Zero256)

		case EXTCODESIZE: // 0x3B
			address := stack.PopAddress()
			if useGasNegative(gas, GasGetAccount, &err) {
				return nil, err
			}
			if callState.Exists(address) {
				code := callState.GetCode(address)
				l := int64(len(code))
				stack.Push64(l)
				vm.Debugf(" => %d\n", l)
			} else {
				if _, ok := registeredNativeContracts[address]; !ok {
					return nil, firstErr(err, errors.ErrorCodeUnknownAddress)
				}
				vm.Debugf(" => returning code size of 1 to indicated existence of native contract at %X\n", address)
				stack.Push(One256)
			}
		case EXTCODECOPY: // 0x3C
			address := stack.PopAddress()
			if useGasNegative(gas, GasGetAccount, &err) {
				return nil, err
			}
			if !callState.Exists(address) {
				if _, ok := registeredNativeContracts[address]; ok {
					vm.Debugf(" => attempted to copy native contract at %v but this is not supported\n", address)
					return nil, firstErr(err, errors.ErrorCodeNativeContractCodeCopy)
				}
				return nil, firstErr(err, errors.ErrorCodeUnknownAddress)
			}
			code := callState.GetCode(address)
			memOff := stack.PopBigInt()
			codeOff, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			length, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			data, ok := subslice(code, codeOff, length)
			if !ok {
				return nil, firstErr(err, errors.ErrorCodeCodeOutOfBounds)
			}
			memErr := memory.Write(memOff, data)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}
			vm.Debugf(" => [%v, %v, %v] %X\n", memOff, codeOff, length, data)

		case RETURNDATASIZE: // 0x3D
			stack.Push64(int64(len(vm.returnData)))
			vm.Debugf(" => %d\n", len(vm.returnData))

		case RETURNDATACOPY: // 0x3E
			memOff, outputOff, length := stack.PopBigInt(), stack.PopBigInt(), stack.PopBigInt()

			end := new(big.Int).Add(outputOff, length)

			if end.BitLen() > 64 || uint64(len(vm.returnData)) < end.Uint64() {
				return nil, errors.ErrorCodeReturnDataOutOfBounds
			}

			data := vm.returnData

			memErr := memory.Write(memOff, data)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}
			vm.Debugf(" => [%v, %v, %v] %X\n", memOff, outputOff, length, data)

		case BLOCKHASH: // 0x40
			stack.Push(Zero256)
			vm.Debugf(" => 0x%X (NOT SUPPORTED)\n", stack.Peek().Bytes())

		case COINBASE: // 0x41
			stack.Push(Zero256)
			vm.Debugf(" => 0x%X (NOT SUPPORTED)\n", stack.Peek().Bytes())

		case TIMESTAMP: // 0x42
			time := vm.params.BlockTime
			stack.Push64(int64(time))
			vm.Debugf(" => 0x%X\n", time)

		case BLOCKHEIGHT: // 0x43
			number := vm.params.BlockHeight
			stack.PushU64(number)
			vm.Debugf(" => 0x%X\n", number)

		case GASLIMIT: // 0x45
			stack.PushU64(vm.params.GasLimit)
			vm.Debugf(" => %v\n", vm.params.GasLimit)

		case POP: // 0x50
			popped := stack.Pop()
			vm.Debugf(" => 0x%X\n", popped)

		case MLOAD: // 0x51
			offset := stack.PopBigInt()
			data, memErr := memory.Read(offset, BigWord256Length)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}
			stack.Push(LeftPadWord256(data))
			vm.Debugf(" => 0x%X @ 0x%X\n", data, offset)

		case MSTORE: // 0x52
			offset, data := stack.PopBigInt(), stack.Pop()
			memErr := memory.Write(offset, data.Bytes())
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}
			vm.Debugf(" => 0x%X @ 0x%X\n", data, offset)

		case MSTORE8: // 0x53
			offset := stack.PopBigInt()
			val64, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			val := byte(val64 & 0xFF)
			memErr := memory.Write(offset, []byte{val})
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}
			vm.Debugf(" => [%v] 0x%X\n", offset, val)

		case SLOAD: // 0x54
			loc := stack.Pop()
			data := callState.GetStorage(callee, loc)
			stack.Push(data)
			vm.Debugf("%s {0x%X = 0x%X}\n", callee, loc, data)

		case SSTORE: // 0x55
			loc, data := stack.Pop(), stack.Pop()
			if useGasNegative(gas, GasStorageUpdate, &err) {
				return nil, err
			}
			callState.SetStorage(callee, loc, data)
			vm.Debugf("%s {0x%X := 0x%X}\n", callee, loc, data)

		case JUMP: // 0x56
			to, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			jumpErr := vm.jump(code, to, &pc)
			if jumpErr != nil {
				vm.Debugf(" => JUMP err: %s", jumpErr)
				return nil, firstErr(err, jumpErr)
			}
			continue

		case JUMPI: // 0x57
			pos, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			cond := stack.Pop()
			if !cond.IsZero() {
				jumpErr := vm.jump(code, pos, &pc)
				if jumpErr != nil {
					return nil, firstErr(err, jumpErr)
				}
				continue
			}
			vm.Debugf(" ~> false\n")

		case PC: // 0x58
			stack.Push64(pc)

		case MSIZE: // 0x59
			// Note: Solidity will write to this offset expecting to find guaranteed
			// free memory to be allocated for it if a subsequent MSTORE is made to
			// this offset.
			capacity := memory.Capacity()
			stack.PushBigInt(capacity)
			vm.Debugf(" => 0x%X\n", capacity)

		case GAS: // 0x5A
			stack.PushU64(*gas)
			vm.Debugf(" => %X\n", *gas)

		case JUMPDEST: // 0x5B
			vm.Debugf("\n")
			// Do nothing

		case PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16, PUSH17, PUSH18, PUSH19, PUSH20, PUSH21, PUSH22, PUSH23, PUSH24, PUSH25, PUSH26, PUSH27, PUSH28, PUSH29, PUSH30, PUSH31, PUSH32:
			a := int64(op - PUSH1 + 1)
			codeSegment, ok := subslice(code, pc+1, a)
			if !ok {
				return nil, firstErr(err, errors.ErrorCodeCodeOutOfBounds)
			}
			res := LeftPadWord256(codeSegment)
			stack.Push(res)
			pc += a
			vm.Debugf(" => 0x%X\n", res)

		case DUP1, DUP2, DUP3, DUP4, DUP5, DUP6, DUP7, DUP8, DUP9, DUP10, DUP11, DUP12, DUP13, DUP14, DUP15, DUP16:
			n := int(op - DUP1 + 1)
			stack.Dup(n)
			vm.Debugf(" => [%d] 0x%X\n", n, stack.Peek().Bytes())

		case SWAP1, SWAP2, SWAP3, SWAP4, SWAP5, SWAP6, SWAP7, SWAP8, SWAP9, SWAP10, SWAP11, SWAP12, SWAP13, SWAP14, SWAP15, SWAP16:
			n := int(op - SWAP1 + 2)
			stack.Swap(n)
			vm.Debugf(" => [%d] %X\n", n, stack.Peek())

		case LOG0, LOG1, LOG2, LOG3, LOG4:
			n := int(op - LOG0)
			topics := make([]Word256, n)
			offset, size := stack.PopBigInt(), stack.PopBigInt()
			for i := 0; i < n; i++ {
				topics[i] = stack.Pop()
			}
			data, memErr := memory.Read(offset, size)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}
			eventErr := eventSink.Log(&exec.LogEvent{
				Address: callee,
				Topics:  topics,
				Data:    data,
			})
			if eventErr != nil {
				return nil, firstErr(err, eventErr)
			}
			vm.Debugf(" => T:%X D:%X\n", topics, data)

		case CREATE: // 0xF0
			vm.returnData = nil

			contractValue, popErr := stack.PopU64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			offset, size := stack.PopBigInt(), stack.PopBigInt()
			input, memErr := memory.Read(offset, size)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}

			// Check balance
			if callState.GetBalance(callee) < uint64(contractValue) {
				return nil, firstErr(err, errors.ErrorCodeInsufficientBalance)
			}

			// TODO charge for gas to create account _ the code length * GasCreateByte
			var gasErr errors.CodedError
			if useGasNegative(gas, GasCreateAccount, &gasErr) {
				return nil, firstErr(err, gasErr)
			}

			callState.IncSequence(callee)
			newAccount := crypto.NewContractAddress(callee, callState.GetSequence(callee))

			// Run the input to get the contract code.
			// NOTE: no need to copy 'input' as per Call contract.
			ret, callErr := vm.Call(callState, eventSink, callee, newAccount, input, input, contractValue, gas)
			if callErr != nil {
				stack.Push(Zero256)
				// Note we both set the return buffer and return the result normally
				vm.returnData = ret
				if callErr.ErrorCode() == errors.ErrorCodeExecutionReverted {
					return ret, callErr
				}
			} else {
				// Set the code (ret need not be copied as per Call contract)
				createErr := createContract(callState, callee, newAccount, ret)
				if createErr != nil {
					return nil, firstErr(err, createErr)
				}
				stack.PushAddress(newAccount)
			}

		case CALL, CALLCODE, DELEGATECALL, STATICCALL: // 0xF1, 0xF2, 0xF4, 0xFA
			vm.returnData = nil

			permErr := EnsurePermission(callState, callee, permission.Call)
			if permErr != nil {
				return nil, firstErr(err, permErr)

			}
			gasLimit, popErr := stack.PopU64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			address := stack.PopAddress()
			// NOTE: for DELEGATECALL value is preserved from the original
			// caller, as such it is not stored on stack as an argument
			// for DELEGATECALL and should not be popped.  Instead previous
			// caller value is used.  for CALL and CALLCODE value is stored
			// on stack and needs to be overwritten from the given value.
			if op != DELEGATECALL && op != STATICCALL {
				value, popErr = stack.PopU64()
				if popErr != nil {
					return nil, firstErr(err, popErr)
				}
			}
			// inputs
			inOffset, inSize := stack.PopBigInt(), stack.PopBigInt()
			// outputs
			retOffset := stack.PopBigInt()
			retSize, popErr := stack.Pop64()
			if popErr != nil {
				return nil, firstErr(err, popErr)
			}
			vm.Debugf(" => %v\n", address)

			// Get the arguments from the memory
			args, memErr := memory.Read(inOffset, inSize)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}

			// Ensure that gasLimit is reasonable
			if *gas < gasLimit {
				// EIP150 - the 63/64 rule - rather than errors.CodedError we pass this specified fraction of the total available gas
				gasLimit = *gas - *gas/64
			}
			// NOTE: we will return any used gas later.
			*gas -= gasLimit

			// Begin execution
			var ret []byte
			var callErr errors.CodedError
			// Establish a stack frame and perform the call
			var childCallState Interface
			if IsRegisteredNativeContract(address) {
				// Native contract
				childCallState = callState.NewCache()
				ret, callErr = ExecuteNativeContract(address, childCallState, callee, args, &gasLimit, logger)
				// for now we fire the Call event. maybe later we'll fire more particulars
				// NOTE: these fire call go_events and not particular go_events for eg name reg or permissions
				vm.fireCallEvent(eventSink, exec.CallTypeSNative, &callErr, &ret, callee, address, args, value,
					&gasLimit, &err)
			} else {
				// EVM contract
				if useGasNegative(gas, GasGetAccount, &callErr) {
					return nil, callErr
				}
				// since CALL is used also for sending funds,
				// acc may not exist yet. This is an errors.CodedError for
				// CALLCODE, but not for CALL, though I don't think
				// ethereum actually cares
				if !callState.Exists(address) {
					if op == CALL {
						// We're sending funds to a new account so we must create it first
						createErr := createAccount(callState, callee, address)
						if createErr != nil {
							return nil, firstErr(err, createErr)
						}
					} else {
						return nil, firstErr(err, errors.ErrorCodeUnknownAddress)
					}
				}
				switch op {
				case CALL:
					childCallState = callState.NewCache()
					ret, callErr = vm.call(childCallState, eventSink, callee, address, callState.GetCode(address),
						args, value, &gasLimit, exec.CallTypeCall)

				case CALLCODE:
					childCallState = callState.NewCache()
					ret, callErr = vm.call(childCallState, eventSink, callee, callee, callState.GetCode(address),
						args, value, &gasLimit, exec.CallTypeCode)

				case DELEGATECALL:
					childCallState = callState.NewCache()
					ret, callErr = vm.delegateCall(childCallState, eventSink, caller, callee,
						callState.GetCode(address), args, value, &gasLimit, exec.CallTypeDelegate)

				case STATICCALL:
					childCallState = callState.NewCache(state.ReadOnly)
					ret, callErr = vm.delegateCall(childCallState, NewLogFreeEventSink(eventSink),
						caller, callee, callState.GetCode(address), args, value, &gasLimit, exec.CallTypeStatic)

				default:
					panic(fmt.Errorf("switch statement should be exhaustive so this should not have been reached"))
				}

			}

			// Set regardless of call error in case execution reverted - does no harm if not reverted
			vm.returnData = ret

			if callErr == nil {
				callErr = childCallState.Sync()
				if callErr != nil {
					return nil, firstErr(err, callErr)
				}
			}

			// Push result
			if callErr != nil {
				vm.Debugf("error from nested sub-call (depth: %v): %s\n", vm.stackDepth, callErr.Error())
				// So we can return nested errors.CodedError if the top level return is an errors.CodedError
				stack.Push(Zero256)

				if callErr.ErrorCode() == errors.ErrorCodeExecutionReverted {
					memory.Write(retOffset, RightPadBytes(ret, int(retSize)))
				}
			} else {
				stack.Push(One256)

				// Should probably only be necessary when there is no return value and
				// ret is empty, but since EVM expects retSize to be respected this will
				// defensively pad or truncate the portion of ret to be returned.
				memErr := memory.Write(retOffset, RightPadBytes(ret, int(retSize)))
				if memErr != nil {
					vm.Debugf(" => Memory err: %s", memErr)
					return nil, firstErr(callErr, errors.ErrorCodeMemoryOutOfBounds)
				}
			}

			// Handle remaining gas.
			*gas += gasLimit

			vm.Debugf("resume %s (%v)\n", callee, gas)

		case RETURN: // 0xF3
			offset, size := stack.PopBigInt(), stack.PopBigInt()
			output, memErr := memory.Read(offset, size)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}
			vm.Debugf(" => [%v, %v] (%d) 0x%X\n", offset, size, len(output), output)
			return output, nil

		case REVERT: // 0xFD
			offset, size := stack.PopBigInt(), stack.PopBigInt()
			output, memErr := memory.Read(offset, size)
			if memErr != nil {
				vm.Debugf(" => Memory err: %s", memErr)
				return nil, firstErr(err, errors.ErrorCodeMemoryOutOfBounds)
			}

			vm.Debugf(" => [%v, %v] (%d) 0x%X\n", offset, size, len(output), output)
			return output, errors.ErrorCodeExecutionReverted

		case INVALID: // 0xFE
			return nil, errors.ErrorCodeExecutionAborted

		case SELFDESTRUCT: // 0xFF
			receiver := stack.PopAddress()
			if useGasNegative(gas, GasGetAccount, &err) {
				return nil, err
			}
			if !callState.Exists(receiver) {
				// If receiver address doesn't exist, try to create it
				var gasErr errors.CodedError
				if useGasNegative(gas, GasCreateAccount, &gasErr) {
					return nil, firstErr(err, gasErr)
				}
				createErr := createAccount(callState, callee, receiver)
				if createErr != nil {
					return nil, firstErr(err, createErr)
				}
			}
			balance := callState.GetBalance(callee)
			callState.AddToBalance(receiver, balance)
			callState.RemoveAccount(callee)
			vm.Debugf(" => (%X) %v\n", receiver[:4], balance)
			fallthrough

		case STOP: // 0x00
			return nil, nil

		case CREATE2:
			return nil, errors.Errorf("%v not yet implemented", op)

		default:
			vm.Debugf("(pc) %-3v Unknown opcode %v\n", pc, op)
			return nil, errors.Errorf("unknown opcode %v", op)
		}
		pc++
	}
}

func createAccount(st Interface, creator, address crypto.Address) errors.CodedError {
	err := EnsurePermission(st, creator, permission.CreateAccount)
	if err != nil {
		return err
	}
	return create(st, address, nil)
}

func createContract(st Interface, creator, address crypto.Address, code []byte) errors.CodedError {
	err := EnsurePermission(st, creator, permission.CreateContract)
	if err != nil {
		return err
	}
	return create(st, address, code)
}

func create(st Interface, address crypto.Address, code []byte) errors.CodedError {
	if IsRegisteredNativeContract(address) {
		return errors.ErrorCodef(errors.ErrorCodeReservedAddress,
			"cannot create account at %v because that address is reserved for a native contract", address)
	}
	st.CreateAccount(address)
	st.InitCode(address, code)
	if st.Error() != nil {
		return errors.Wrap(st.Error(), "createAccount could not create account")
	}
	return nil
}

// TODO: [Silas] this function seems extremely dubious to me. It was being used
// in circumstances where its behaviour did not match the intention. It's bounds
// check is strange (treats a read at data length as a zero read of arbitrary length)
// I have left it in for now to be conservative about where its behaviour is being used
//
// Returns a subslice from offset of length length and a bool
// (true iff slice was possible). If the subslice
// extends past the end of data it returns A COPY of the segment at the end of
// data padded with zeroes on the right. If offset == len(data) it returns all
// zeroes. if offset > len(data) it returns a false
func subslice(data []byte, offset, length int64) (ret []byte, ok bool) {
	size := int64(len(data))
	if size < offset || offset < 0 || length < 0 {
		return nil, false
	} else if size < offset+length {
		ret, ok = data[offset:], true
		ret = RightPadBytes(ret, 32)
	} else {
		ret, ok = data[offset:offset+length], true
	}
	return
}

func codeGetOp(code []byte, n int64) OpCode {
	if int64(len(code)) <= n {
		return OpCode(0) // stop
	} else {
		return OpCode(code[n])
	}
}

func (vm *VM) jump(code []byte, to int64, pc *int64) (err errors.CodedError) {
	dest := codeGetOp(code, to)
	if dest != JUMPDEST {
		vm.Debugf(" ~> %v invalid jump dest %v\n", to, dest)
		return errors.ErrorCodeInvalidJumpDest
	}
	vm.Debugf(" ~> %v\n", to)
	*pc = to
	return nil
}

func firstErr(errA, errB error) errors.CodedError {
	if errA != nil {
		return errors.AsException(errA)
	} else {
		return errors.AsException(errB)
	}
}

func transfer(st Interface, from, to crypto.Address, amount uint64) errors.CodedError {
	if amount == 0 {
		return nil
	}
	if st.GetBalance(from) < amount {
		return errors.ErrorCodeInsufficientBalance
	} else {
		st.SubtractFromBalance(from, amount)
		st.AddToBalance(to, amount)
	}
	err := st.Error()
	if err != nil {
		return err
	}
	return nil
}

// Dump the bytecode being sent to the EVM in the current working directory
func dumpTokens(txHash []byte, caller, callee crypto.Address, code []byte) {
	var tokensString string
	tokens, err := acm.Bytecode(code).Tokens()
	if err != nil {
		tokensString = fmt.Sprintf("error generating tokens from bytecode: %v", err)
	} else {
		tokensString = strings.Join(tokens, "\n")
	}
	txHashString := "tx-none"
	if len(txHash) >= 4 {
		txHashString = fmt.Sprintf("tx-%X", txHash[:4])
	}
	callerString := "caller-none"
	if caller != crypto.ZeroAddress {
		callerString = fmt.Sprintf("caller-%v", caller)
	}
	calleeString := "callee-none"
	if callee != crypto.ZeroAddress {
		calleeString = fmt.Sprintf("callee-%v", caller)
	}
	ioutil.WriteFile(fmt.Sprintf("tokens_%s_%s_%s.asm", txHashString, callerString, calleeString),
		[]byte(tokensString), 0777)
}

func (vm *VM) ensureStackDepth() errors.CodedError {
	if vm.params.CallStackMaxDepth > 0 && vm.stackDepth == vm.params.CallStackMaxDepth {
		return errors.ErrorCodeCallStackOverflow
	}
	return nil
}
