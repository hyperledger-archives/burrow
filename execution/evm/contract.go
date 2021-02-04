package evm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"

	"github.com/hyperledger/burrow/acm"
	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	. "github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
)

type Contract struct {
	*EVM
	*Code
}

func (c *Contract) Call(state engine.State, params engine.CallParams) ([]byte, error) {
	return engine.Call(state, params, c.execute)
}

// Executes the EVM code passed in the appropriate context
func (c *Contract) execute(st engine.State, params engine.CallParams) ([]byte, error) {
	c.debugf("(%d) (%s) %s (code=%d) gas: %v (d) %X\n",
		st.CallFrame.CallStackDepth(), params.Caller, params.Callee, c.Length(), *params.Gas, params.Input)

	if c.Length() == 0 {
		return nil, nil
	}

	if c.options.DumpTokens {
		dumpTokens(c.options.Nonce, params.Caller, params.Callee, c.GetBytecode())
	}

	// Program counter - the index into code that tracks current instruction
	var pc uint64
	// Return data from a call
	var returnData []byte

	// Maybe serves 3 purposes: 1. provides 'capture first error semantics', 2. reduces clutter of error handling
	// particular for 1, 3. acts a shared error sink for stack, memory, and the main execute loop
	maybe := new(errors.Maybe)

	// Provide stack and memory storage - passing in the callState as an error provider
	stack := NewStack(maybe, c.options.DataStackInitialCapacity, c.options.DataStackMaxDepth, params.Gas)
	memory := c.options.MemoryProvider(maybe)

	for {
		// Check for any error in this frame.
		if maybe.Error() != nil {
			return nil, maybe.Error()
		}

		var op = c.GetSymbol(pc)
		c.debugf("(pc) %-3d (op) %-14s (st) %-4d (gas) %d", pc, op.String(), stack.Len(), params.Gas)
		// Use BaseOp gas.
		maybe.PushError(engine.UseGasNegative(params.Gas, engine.GasBaseOp))

		switch op {

		case ADD: // 0x01
			x, y := stack.PopBigInt(), stack.PopBigInt()
			sum := new(big.Int).Add(x, y)
			res := stack.PushBigInt(sum)
			c.debugf(" %v + %v = %v (%v)\n", x, y, sum, res)

		case MUL: // 0x02
			x, y := stack.PopBigInt(), stack.PopBigInt()
			prod := new(big.Int).Mul(x, y)
			res := stack.PushBigInt(prod)
			c.debugf(" %v * %v = %v (%v)\n", x, y, prod, res)

		case SUB: // 0x03
			x, y := stack.PopBigInt(), stack.PopBigInt()
			diff := new(big.Int).Sub(x, y)
			res := stack.PushBigInt(diff)
			c.debugf(" %v - %v = %v (%v)\n", x, y, diff, res)

		case DIV: // 0x04
			x, y := stack.PopBigInt(), stack.PopBigInt()
			if y.Sign() == 0 {
				stack.Push(Zero256)
				c.debugf(" %v / %v = %v\n", x, y, 0)
			} else {
				div := new(big.Int).Div(x, y)
				res := stack.PushBigInt(div)
				c.debugf(" %v / %v = %v (%v)\n", x, y, div, res)
			}

		case SDIV: // 0x05
			x, y := stack.PopBigIntSigned(), stack.PopBigIntSigned()
			if y.Sign() == 0 {
				stack.Push(Zero256)
				c.debugf(" %v / %v = %v\n", x, y, 0)
			} else {
				div := new(big.Int).Quo(x, y)
				res := stack.PushBigInt(div)
				c.debugf(" %v / %v = %v (%v)\n", x, y, div, res)
			}

		case MOD: // 0x06
			x, y := stack.PopBigInt(), stack.PopBigInt()
			if y.Sign() == 0 {
				stack.Push(Zero256)
				c.debugf(" %v %% %v = %v\n", x, y, 0)
			} else {
				mod := new(big.Int).Mod(x, y)
				res := stack.PushBigInt(mod)
				c.debugf(" %v %% %v = %v (%v)\n", x, y, mod, res)
			}

		case SMOD: // 0x07
			x, y := stack.PopBigIntSigned(), stack.PopBigIntSigned()
			if y.Sign() == 0 {
				stack.Push(Zero256)
				c.debugf(" %v %% %v = %v\n", x, y, 0)
			} else {
				mod := new(big.Int).Rem(x, y)
				res := stack.PushBigInt(mod)
				c.debugf(" %v %% %v = %v (%v)\n", x, y, mod, res)
			}

		case ADDMOD: // 0x08
			x, y, z := stack.PopBigInt(), stack.PopBigInt(), stack.PopBigInt()
			if z.Sign() == 0 {
				stack.Push(Zero256)
				c.debugf(" %v %% %v = %v\n", x, y, 0)
			} else {
				add := new(big.Int).Add(x, y)
				mod := add.Mod(add, z)
				res := stack.PushBigInt(mod)
				c.debugf(" %v + %v %% %v = %v (%v)\n", x, y, z, mod, res)
			}

		case MULMOD: // 0x09
			x, y, z := stack.PopBigInt(), stack.PopBigInt(), stack.PopBigInt()
			if z.Sign() == 0 {
				stack.Push(Zero256)
				c.debugf(" %v %% %v = %v\n", x, y, 0)
			} else {
				mul := new(big.Int).Mul(x, y)
				mod := mul.Mod(mul, z)
				res := stack.PushBigInt(mod)
				c.debugf(" %v * %v %% %v = %v (%v)\n", x, y, z, mod, res)
			}

		case EXP: // 0x0A
			x, y := stack.PopBigInt(), stack.PopBigInt()
			pow := new(big.Int).Exp(x, y, nil)
			res := stack.PushBigInt(pow)
			c.debugf(" %v ** %v = %v (%v)\n", x, y, pow, res)

		case SIGNEXTEND: // 0x0B
			back := stack.PopBigInt().Uint64()
			if back < Word256Bytes-1 {
				bits := uint((back + 1) * 8)
				stack.PushBigInt(SignExtend(stack.PopBigInt(), bits))
			}
			// Continue leaving the sign extension argument on the stack. This makes sign-extending a no-op if embedded
			// integer is already one word wide

		case LT: // 0x10
			x, y := stack.PopBigInt(), stack.PopBigInt()
			if x.Cmp(y) < 0 {
				stack.Push(One256)
				c.debugf(" %v < %v = %v\n", x, y, 1)
			} else {
				stack.Push(Zero256)
				c.debugf(" %v < %v = %v\n", x, y, 0)
			}

		case GT: // 0x11
			x, y := stack.PopBigInt(), stack.PopBigInt()
			if x.Cmp(y) > 0 {
				stack.Push(One256)
				c.debugf(" %v > %v = %v\n", x, y, 1)
			} else {
				stack.Push(Zero256)
				c.debugf(" %v > %v = %v\n", x, y, 0)
			}

		case SLT: // 0x12
			x, y := stack.PopBigIntSigned(), stack.PopBigIntSigned()
			if x.Cmp(y) < 0 {
				stack.Push(One256)
				c.debugf(" %v < %v = %v\n", x, y, 1)
			} else {
				stack.Push(Zero256)
				c.debugf(" %v < %v = %v\n", x, y, 0)
			}

		case SGT: // 0x13
			x, y := stack.PopBigIntSigned(), stack.PopBigIntSigned()
			if x.Cmp(y) > 0 {
				stack.Push(One256)
				c.debugf(" %v > %v = %v\n", x, y, 1)
			} else {
				stack.Push(Zero256)
				c.debugf(" %v > %v = %v\n", x, y, 0)
			}

		case EQ: // 0x14
			x, y := stack.Pop(), stack.Pop()
			if bytes.Equal(x[:], y[:]) {
				stack.Push(One256)
				c.debugf(" %v == %v = %v\n", x, y, 1)
			} else {
				stack.Push(Zero256)
				c.debugf(" %v == %v = %v\n", x, y, 0)
			}

		case ISZERO: // 0x15
			x := stack.Pop()
			if x.IsZero() {
				stack.Push(One256)
				c.debugf(" %v == 0 = %v\n", x, 1)
			} else {
				stack.Push(Zero256)
				c.debugf(" %v == 0 = %v\n", x, 0)
			}

		case AND: // 0x16
			x, y := stack.Pop(), stack.Pop()
			z := [32]byte{}
			for i := 0; i < 32; i++ {
				z[i] = x[i] & y[i]
			}
			stack.Push(z)
			c.debugf(" %v & %v = %v\n", x, y, z)

		case OR: // 0x17
			x, y := stack.Pop(), stack.Pop()
			z := [32]byte{}
			for i := 0; i < 32; i++ {
				z[i] = x[i] | y[i]
			}
			stack.Push(z)
			c.debugf(" %v | %v = %v\n", x, y, z)

		case XOR: // 0x18
			x, y := stack.Pop(), stack.Pop()
			z := [32]byte{}
			for i := 0; i < 32; i++ {
				z[i] = x[i] ^ y[i]
			}
			stack.Push(z)
			c.debugf(" %v ^ %v = %v\n", x, y, z)

		case NOT: // 0x19
			x := stack.Pop()
			z := [32]byte{}
			for i := 0; i < 32; i++ {
				z[i] = ^x[i]
			}
			stack.Push(z)
			c.debugf(" !%v = %v\n", x, z)

		case BYTE: // 0x1A
			idx := stack.Pop64()
			val := stack.Pop()
			res := byte(0)
			if idx < 32 {
				res = val[idx]
			}
			stack.Push64(uint64(res))
			c.debugf(" => 0x%X\n", res)

		case SHL: //0x1B
			shift, x := stack.PopBigInt(), stack.PopBigInt()

			if shift.Cmp(Big256) >= 0 {
				reset := big.NewInt(0)
				stack.PushBigInt(reset)
				c.debugf(" %v << %v = %v\n", x, shift, reset)
			} else {
				shiftedValue := x.Lsh(x, uint(shift.Uint64()))
				stack.PushBigInt(shiftedValue)
				c.debugf(" %v << %v = %v\n", x, shift, shiftedValue)
			}

		case SHR: //0x1C
			shift, x := stack.PopBigInt(), stack.PopBigInt()

			if shift.Cmp(Big256) >= 0 {
				reset := big.NewInt(0)
				stack.PushBigInt(reset)
				c.debugf(" %v << %v = %v\n", x, shift, reset)
			} else {
				shiftedValue := x.Rsh(x, uint(shift.Uint64()))
				stack.PushBigInt(shiftedValue)
				c.debugf(" %v << %v = %v\n", x, shift, shiftedValue)
			}

		case SAR: //0x1D
			shift, x := stack.PopBigInt(), stack.PopBigIntSigned()

			if shift.Cmp(Big256) >= 0 {
				reset := big.NewInt(0)
				if x.Sign() < 0 {
					reset.SetInt64(-1)
				}
				stack.PushBigInt(reset)
				c.debugf(" %v << %v = %v\n", x, shift, reset)
			} else {
				shiftedValue := x.Rsh(x, uint(shift.Uint64()))
				stack.PushBigInt(shiftedValue)
				c.debugf(" %v << %v = %v\n", x, shift, shiftedValue)
			}

		case SHA3: // 0x20
			maybe.PushError(engine.UseGasNegative(params.Gas, engine.GasSha3))
			offset, size := stack.PopBigInt(), stack.PopBigInt()
			data := memory.Read(offset, size)
			data = crypto.Keccak256(data)
			stack.PushBytes(data)
			c.debugf(" => (%v) %X\n", size, data)

		case ADDRESS: // 0x30
			stack.Push(params.Callee.Word256())
			c.debugf(" => %v\n", params.Callee)

		case BALANCE: // 0x31
			address := stack.PopAddress()
			maybe.PushError(engine.UseGasNegative(params.Gas, engine.GasGetAccount))
			balance := engine.MustGetAccount(st.CallFrame, maybe, address).Balance
			stack.Push64(balance)
			c.debugf(" => %v (%v)\n", balance, address)

		case ORIGIN: // 0x32
			stack.Push(params.Origin.Word256())
			c.debugf(" => %v\n", params.Origin)

		case CALLER: // 0x33
			stack.Push(params.Caller.Word256())
			c.debugf(" => %v\n", params.Caller)

		case CALLVALUE: // 0x34
			stack.PushBigInt(&params.Value)
			c.debugf(" => %v\n", params.Value)

		case CALLDATALOAD: // 0x35
			offset := stack.Pop64()
			data := maybe.Bytes(subslice(params.Input, offset, 32))
			res := LeftPadWord256(data)
			stack.Push(res)
			c.debugf(" => 0x%v\n", res)

		case CALLDATASIZE: // 0x36
			stack.Push64(uint64(len(params.Input)))
			c.debugf(" => %d\n", len(params.Input))

		case CALLDATACOPY: // 0x37
			memOff := stack.PopBigInt()
			inputOff := stack.Pop64()
			length := stack.Pop64()
			data := maybe.Bytes(subslice(params.Input, inputOff, length))
			memory.Write(memOff, data)
			c.debugf(" => [%v, %v, %v] %X\n", memOff, inputOff, length, data)

		case CODESIZE: // 0x38
			l := uint64(c.Length())
			stack.Push64(l)
			c.debugf(" => %d\n", l)

		case CODECOPY: // 0x39
			memOff := stack.PopBigInt()
			codeOff := stack.Pop64()
			length := stack.Pop64()
			data := maybe.Bytes(subslice(c.GetBytecode(), codeOff, length))
			memory.Write(memOff, data)
			c.debugf(" => [%v, %v, %v] %X\n", memOff, codeOff, length, data)

		case GASPRICE_DEPRECATED: // 0x3A
			stack.Push(Zero256)
			c.debugf(" => %v (GASPRICE IS DEPRECATED)\n", Zero256)

		case EXTCODESIZE: // 0x3B
			address := stack.PopAddress()
			maybe.PushError(engine.UseGasNegative(params.Gas, engine.GasGetAccount))
			acc := engine.MustGetAccount(st.CallFrame, maybe, address)
			if acc == nil {
				stack.Push(Zero256)
				c.debugf(" => 0\n")
			} else {
				length := uint64(len(acc.Code()))
				stack.Push64(length)
				c.debugf(" => %d\n", length)
			}
		case EXTCODECOPY: // 0x3C
			address := stack.PopAddress()
			maybe.PushError(engine.UseGasNegative(params.Gas, engine.GasGetAccount))
			acc := engine.MustGetAccount(st.CallFrame, maybe, address)
			if acc == nil {
				maybe.PushError(errors.Codes.UnknownAddress)
			} else {
				code := acc.EVMCode
				memOff := stack.PopBigInt()
				codeOff := stack.Pop64()
				length := stack.Pop64()
				data := maybe.Bytes(subslice(code, codeOff, length))
				memory.Write(memOff, data)
				c.debugf(" => [%v, %v, %v] %X\n", memOff, codeOff, length, data)
			}

		case RETURNDATASIZE: // 0x3D
			stack.Push64(uint64(len(returnData)))
			c.debugf(" => %d\n", len(returnData))

		case RETURNDATACOPY: // 0x3E
			memOff, outputOff, length := stack.PopBigInt(), stack.PopBigInt(), stack.PopBigInt()
			end := new(big.Int).Add(outputOff, length)

			if end.BitLen() > 64 || uint64(len(returnData)) < end.Uint64() {
				maybe.PushError(errors.Codes.ReturnDataOutOfBounds)
				continue
			}

			memory.Write(memOff, returnData)
			c.debugf(" => [%v, %v, %v] %X\n", memOff, outputOff, length, returnData)

		case EXTCODEHASH: // 0x3F
			address := stack.PopAddress()

			acc := engine.GetAccount(st.CallFrame, maybe, address)
			if acc == nil {
				// In case the account does not exist 0 is pushed to the stack.
				stack.Push64(0)
			} else {
				// keccak256 hash of a contract's code
				var extcodehash Word256
				if len(acc.CodeHash) > 0 {
					copy(extcodehash[:], acc.CodeHash)
				} else {
					copy(extcodehash[:], crypto.Keccak256(acc.Code()))
				}
				stack.Push(extcodehash)
			}

		case BLOCKHASH: // 0x40
			blockNumber := stack.Pop64()

			lastBlockHeight := st.Blockchain.LastBlockHeight()
			if blockNumber >= lastBlockHeight {
				c.debugf(" => attempted to get block hash of a non-existent block: %v", blockNumber)
				maybe.PushError(errors.Codes.InvalidBlockNumber)
			} else if lastBlockHeight-blockNumber > MaximumAllowedBlockLookBack {
				c.debugf(" => attempted to get block hash of a block %d outside of the allowed range "+
					"(must be within %d blocks)", blockNumber, MaximumAllowedBlockLookBack)
				maybe.PushError(errors.Codes.BlockNumberOutOfRange)
			} else {
				hash := maybe.Bytes(st.Blockchain.BlockHash(blockNumber))
				blockHash := LeftPadWord256(hash)
				stack.Push(blockHash)
				c.debugf(" => 0x%v\n", blockHash)
			}

		case COINBASE: // 0x41
			stack.Push(Zero256)
			c.debugf(" => 0x%v (NOT SUPPORTED)\n", stack.Peek())

		case TIMESTAMP: // 0x42
			blockTime := st.Blockchain.LastBlockTime().Unix()
			stack.Push64(uint64(blockTime))
			c.debugf(" => %d\n", blockTime)

		case BLOCKHEIGHT: // 0x43
			number := st.Blockchain.LastBlockHeight()
			stack.Push64(number)
			c.debugf(" => %d\n", number)

		case DIFFICULTY: // 0x44
			// ~ hashes per solution - by convention we'll use unity since there are no misses if you are proposer
			stack.Push(One256)
			c.debugf(" => %v\n", One256)

		case GASLIMIT: // 0x45
			stack.PushBigInt(params.Gas)
			c.debugf(" => %v\n", *params.Gas)

		case CHAINID: // 0x46
			id := crypto.GetEthChainID(st.Blockchain.ChainID())
			stack.PushBigInt(id)
			c.debugf(" => %X\n", id)

		case POP: // 0x50
			popped := stack.Pop()
			c.debugf(" => 0x%v\n", popped)

		case MLOAD: // 0x51
			offset := stack.PopBigInt()
			data := memory.Read(offset, BigWord256Bytes)
			stack.Push(LeftPadWord256(data))
			c.debugf(" => 0x%X @ 0x%v\n", data, offset)

		case MSTORE: // 0x52
			offset, data := stack.PopBigInt(), stack.Pop()
			memory.Write(offset, data.Bytes())
			c.debugf(" => 0x%v @ 0x%v\n", data, offset)

		case MSTORE8: // 0x53
			offset := stack.PopBigInt()
			val64 := stack.PopBigInt().Uint64()
			val := byte(val64 & 0xFF)
			memory.Write(offset, []byte{val})
			c.debugf(" => [%v] 0x%X\n", offset, val)

		case SLOAD: // 0x54
			loc := stack.Pop()
			data := LeftPadWord256(maybe.Bytes(st.CallFrame.GetStorage(params.Callee, loc)))
			stack.Push(data)
			c.debugf("%v {0x%v = 0x%v}\n", params.Callee, loc, data)

		case SSTORE: // 0x55
			loc, data := stack.Pop(), stack.Pop()
			maybe.PushError(engine.UseGasNegative(params.Gas, engine.GasStorageUpdate))
			maybe.PushError(st.CallFrame.SetStorage(params.Callee, loc, data.Bytes()))
			c.debugf("%v {%v := %v}\n", params.Callee, loc, data)

		case JUMP: // 0x56
			to := stack.Pop64()
			maybe.PushError(c.jump(to, &pc))
			continue

		case JUMPI: // 0x57
			pos := stack.Pop64()
			cond := stack.Pop()
			if !cond.IsZero() {
				maybe.PushError(c.jump(pos, &pc))
				continue
			} else {
				c.debugf(" ~> false\n")
			}

		case PC: // 0x58
			stack.Push64(pc)

		case MSIZE: // 0x59
			// Note: Solidity will write to this offset expecting to find guaranteed
			// free memory to be allocated for it if a subsequent MSTORE is made to
			// this offset.
			capacity := memory.Capacity()
			stack.PushBigInt(capacity)
			c.debugf(" => 0x%X\n", capacity)

		case GAS: // 0x5A
			stack.PushBigInt(params.Gas)
			c.debugf(" => %X\n", params.Gas)

		case JUMPDEST: // 0x5B
			c.debugf("\n")
			// Do nothing

		case PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16, PUSH17, PUSH18, PUSH19, PUSH20, PUSH21, PUSH22, PUSH23, PUSH24, PUSH25, PUSH26, PUSH27, PUSH28, PUSH29, PUSH30, PUSH31, PUSH32:
			a := uint64(op - PUSH1 + 1)
			codeSegment := maybe.Bytes(subslice(c.GetBytecode(), pc+1, a))
			res := LeftPadWord256(codeSegment)
			stack.Push(res)
			pc += a
			c.debugf(" => 0x%v\n", res)

		case DUP1, DUP2, DUP3, DUP4, DUP5, DUP6, DUP7, DUP8, DUP9, DUP10, DUP11, DUP12, DUP13, DUP14, DUP15, DUP16:
			n := int(op - DUP1 + 1)
			stack.Dup(n)
			c.debugf(" => [%d] 0x%v\n", n, stack.Peek())

		case SWAP1, SWAP2, SWAP3, SWAP4, SWAP5, SWAP6, SWAP7, SWAP8, SWAP9, SWAP10, SWAP11, SWAP12, SWAP13, SWAP14, SWAP15, SWAP16:
			n := int(op - SWAP1 + 2)
			stack.Swap(n)
			c.debugf(" => [%d] %v\n", n, stack.Peek())

		case LOG0, LOG1, LOG2, LOG3, LOG4:
			n := int(op - LOG0)
			topics := make([]Word256, n)
			offset, size := stack.PopBigInt(), stack.PopBigInt()
			for i := 0; i < n; i++ {
				topics[i] = stack.Pop()
			}
			data := memory.Read(offset, size)
			maybe.PushError(st.EventSink.Log(&exec.LogEvent{
				Address: params.Callee,
				Topics:  topics,
				Data:    data,
			}))
			c.debugf(" => T:%v D:%X\n", topics, data)

		case CREATE, CREATE2: // 0xF0, 0xFB
			returnData = nil
			contractValue := stack.PopBigInt()
			offset, size := stack.PopBigInt(), stack.PopBigInt()
			input := memory.Read(offset, size)

			// TODO charge for gas to create account _ the code length * GasCreateByte
			maybe.PushError(engine.UseGasNegative(params.Gas, engine.GasCreateAccount))

			var newAccountAddress crypto.Address
			if op == CREATE {
				c.sequence++
				nonce := make([]byte, txs.HashLength+uint64Length)
				copy(nonce, c.options.Nonce)
				binary.BigEndian.PutUint64(nonce[txs.HashLength:], c.sequence)
				newAccountAddress = crypto.NewContractAddress(params.Callee, nonce)
			} else if op == CREATE2 {
				salt := stack.Pop()
				code := engine.MustGetAccount(st.CallFrame, maybe, params.Callee).EVMCode
				newAccountAddress = crypto.NewContractAddress2(params.Callee, salt, code)
			}

			// Check the CreateContract permission for this account
			if maybe.PushError(engine.EnsurePermission(st.CallFrame, params.Callee, permission.CreateContract)) {
				continue
			}

			// Establish a frame in which the putative account exists
			childCallFrame, err := st.CallFrame.NewFrame()
			maybe.PushError(err)
			maybe.PushError(engine.CreateAccount(childCallFrame, newAccountAddress))

			// Run the input to get the contract code.
			// NOTE: no need to copy 'input' as per Call contract.
			ret, callErr := c.Contract(input).Call(
				engine.State{
					CallFrame:  childCallFrame,
					Blockchain: st.Blockchain,
					EventSink:  st.EventSink,
				},
				engine.CallParams{
					Origin: params.Origin,
					Caller: params.Callee,
					Callee: newAccountAddress,
					Input:  input,
					Value:  *contractValue,
					Gas:    params.Gas,
				})
			if callErr != nil {
				stack.Push(Zero256)
				// Note we both set the return buffer and return the result normally in order to service the error to
				// EVM caller
				returnData = ret
			} else {
				// Update the account with its initialised contract code
				maybe.PushError(engine.InitChildCode(childCallFrame, newAccountAddress, params.Callee, ret))
				maybe.PushError(childCallFrame.Sync())
				stack.PushAddress(newAccountAddress)
			}

		case CALL, CALLCODE, DELEGATECALL, STATICCALL: // 0xF1, 0xF2, 0xF4, 0xFA
			returnData = nil

			if maybe.PushError(engine.EnsurePermission(st.CallFrame, params.Callee, permission.Call)) {
				continue
			}
			// Pull arguments off stack:
			gasLimit := stack.PopBigInt()
			target := stack.PopAddress()
			value := params.Value
			// NOTE: for DELEGATECALL value is preserved from the original
			// caller, as such it is not stored on stack as an argument
			// for DELEGATECALL and should not be popped.  Instead previous
			// caller value is used.  for CALL and CALLCODE value is stored
			// on stack and needs to be overwritten from the given value.
			if op != DELEGATECALL && op != STATICCALL {
				value = *stack.PopBigInt()
			}
			// inputs
			inOffset, inSize := stack.PopBigInt(), stack.PopBigInt()
			// outputs
			retOffset := stack.PopBigInt()
			retSize := stack.Pop64()
			c.debugf(" => %v\n", target)

			var err error
			returnData, err = engine.CallFromSite(st, c.externalDispatcher, params, engine.CallParams{
				CallType: callTypeFromOpCode(op),
				Callee:   target,
				Input:    memory.Read(inOffset, inSize),
				Value:    value,
				Gas:      gasLimit,
			})
			// Push result
			if err != nil {
				c.debugf("error from nested sub-call (depth: %v): %s\n", st.CallFrame.CallStackDepth(), err)
				// So we can return nested errors.CodedError if the top level return is an errors.CodedError
				stack.Push(Zero256)

			} else {
				stack.Push(One256)
			}

			code := errors.GetCode(err)
			if code == errors.Codes.None || code == errors.Codes.ExecutionReverted {
				memory.Write(retOffset, RightPadBytes(returnData, int(retSize)))
			} else {
				maybe.PushError(err)
			}

			// TODO: decide how to handle this
			// Apply refund of any unused gas
			params.Gas.Add(params.Gas, gasLimit)

			c.debugf("resume %s (%v)\n", params.Callee, params.Gas)

		case RETURN: // 0xF3
			offset, size := stack.PopBigInt(), stack.PopBigInt()
			output := memory.Read(offset, size)
			c.debugf(" => [%v, %v] (%d) 0x%X\n", offset, size, len(output), output)
			return output, maybe.Error()

		case REVERT: // 0xFD
			offset, size := stack.PopBigInt(), stack.PopBigInt()
			output := memory.Read(offset, size)
			c.debugf(" => [%v, %v] (%d) 0x%X\n", offset, size, len(output), output)
			maybe.PushError(newRevertException(output))
			return output, maybe.Error()

		case INVALID: // 0xFE
			maybe.PushError(errors.Codes.ExecutionAborted)
			return nil, maybe.Error()

		case SELFDESTRUCT: // 0xFF
			receiver := stack.PopAddress()
			maybe.PushError(engine.UseGasNegative(params.Gas, engine.GasGetAccount))
			if engine.GetAccount(st.CallFrame, maybe, receiver) == nil {
				// If receiver address doesn't exist, try to create it
				maybe.PushError(engine.UseGasNegative(params.Gas, engine.GasCreateAccount))
				if maybe.PushError(st.CallFrame.CreateAccount(params.Callee, receiver)) {
					continue
				}
			}
			balance := engine.MustGetAccount(st.CallFrame, maybe, params.Callee).Balance
			maybe.PushError(engine.UpdateAccount(st.CallFrame, receiver, func(account *acm.Account) error {
				return account.AddToBalance(balance)
			}))
			maybe.PushError(engine.RemoveAccount(st.CallFrame, params.Callee))
			c.debugf(" => (%X) %v\n", receiver[:4], balance)
			return nil, maybe.Error()

		case STOP: // 0x00
			c.debugf("\n")
			return nil, maybe.Error()

		default:
			c.debugf("(pc) %-3v Unknown opcode %v\n", pc, op)
			maybe.PushError(errors.Errorf(errors.Codes.Generic, "unknown opcode %v", op))
			return nil, maybe.Error()
		}
		pc++
	}
	return nil, maybe.Error()
}

func (c *Contract) jump(to uint64, pc *uint64) error {
	dest := c.GetSymbol(to)
	if dest != JUMPDEST || c.IsPushData(to) {
		c.debugf(" ~> %v invalid jump dest %v\n", to, dest)
		return errors.Codes.InvalidJumpDest
	}
	c.debugf(" ~> %v\n", to)
	*pc = to
	return nil
}

// Returns a subslice from offset of length length and a bool
// (true iff slice was possible). If the subslice
// extends past the end of data it returns A COPY of the segment at the end of
// data padded with zeroes on the right. If offset == len(data) it returns all
// zeroes. if offset > len(data) it returns a false
func subslice(data []byte, offset, length uint64) ([]byte, error) {
	size := uint64(len(data))
	if size < offset || offset < 0 || length < 0 {
		return nil, errors.Errorf(errors.Codes.InputOutOfBounds,
			"subslice could not slice data of size %d at offset %d for length %d", size, offset, length)
	}
	if size < offset+length {
		// Extract slice from offset to end padding to requested length
		ret := make([]byte, length)
		copy(ret, data[offset:])
		return ret, nil
	}
	return data[offset : offset+length], nil
}

// Dump the bytecode being sent to the EVM in the current working directory
func dumpTokens(nonce []byte, caller, callee crypto.Address, code []byte) {
	var tokensString string
	tokens, err := acm.Bytecode(code).Tokens()
	if err != nil {
		tokensString = fmt.Sprintf("error generating tokens from bytecode: %v", err)
	} else {
		tokensString = strings.Join(tokens, "\n")
	}
	txHashString := "nil-nonce"
	if len(nonce) >= 4 {
		txHashString = fmt.Sprintf("nonce-%X", nonce[:4])
	}
	callerString := "caller-none"
	if caller != crypto.ZeroAddress {
		callerString = fmt.Sprintf("caller-%v", caller)
	}
	calleeString := "callee-none"
	if callee != crypto.ZeroAddress {
		calleeString = fmt.Sprintf("callee-%v", caller)
	}
	_ = ioutil.WriteFile(fmt.Sprintf("tokens_%s_%s_%s.asm", txHashString, callerString, calleeString),
		[]byte(tokensString), 0777)
}

func newRevertException(ret []byte) errors.CodedError {
	code := errors.Codes.ExecutionReverted
	if len(ret) > 0 {
		// Attempt decode
		reason, err := abi.UnpackRevert(ret)
		if err == nil {
			return errors.Errorf(code, "with reason '%s'", *reason)
		}
	}
	return code
}

func callTypeFromOpCode(o OpCode) exec.CallType {
	switch o {
	case CALL:
		return exec.CallTypeCall
	case CALLCODE:
		return exec.CallTypeCode
	case STATICCALL:
		return exec.CallTypeStatic
	case DELEGATECALL:
		return exec.CallTypeDelegate
	default:
		return exec.CallTypeInvalid
	}
}
