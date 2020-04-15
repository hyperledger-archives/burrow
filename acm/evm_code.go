package acm

import (
	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	"github.com/tmthrgd/go-bitset"
)

func MustEVMCodeFrom(bytelikes ...interface{}) *EVMCode {
	code, err := EVMCodeFrom(bytelikes...)
	if err != nil {
		panic(err)
	}
	return code
}

// Builds new bytecode using the Splice helper to map byte-like and byte-slice-like types to a flat bytecode slice
func EVMCodeFrom(bytelikes ...interface{}) (*EVMCode, error) {
	code, err := bc.Splice(bytelikes...)
	if err != nil {
		return nil, err
	}
	return NewEVMCode(code), nil
}

func NewEVMCode(code []byte) *EVMCode {
	return &EVMCode{
		Bytecode:     code,
		OpcodeBitset: EVMOpcodeBitset(code),
	}
}

func (evmCode *EVMCode) Length() int {
	if evmCode == nil {
		return 0
	}
	return len(evmCode.Bytecode)
}

func (evmCode *EVMCode) GetBytecode() Bytecode {
	if evmCode == nil {
		return nil
	}
	return evmCode.Bytecode
}

func (evmCode *EVMCode) IsOpcode(indexOfSymbolInCode uint64) bool {
	if evmCode == nil || indexOfSymbolInCode >= uint64(evmCode.OpcodeBitset.Len()) {
		return false
	}
	return evmCode.OpcodeBitset.IsSet(uint(indexOfSymbolInCode))
}

func (evmCode *EVMCode) IsPushData(indexOfSymbolInCode uint64) bool {
	return !evmCode.IsOpcode(indexOfSymbolInCode)
}

func (evmCode *EVMCode) GetSymbol(n uint64) asm.OpCode {
	if uint64(evmCode.Length()) <= n {
		return asm.OpCode(0) // stop
	} else {
		return asm.OpCode(evmCode.Bytecode[n])
	}
}

// If code[i] is an opcode (rather than PUSH data) then bitset.IsSet(i) will be true
func EVMOpcodeBitset(code []byte) bitset.Bitset {
	bs := bitset.New(uint(len(code)))
	for i := 0; i < len(code); i++ {
		bs.Set(uint(i))
		symbol := asm.OpCode(code[i])
		if symbol >= asm.PUSH1 && symbol <= asm.PUSH32 {
			i += int(symbol - asm.PUSH1 + 1)
		}
	}
	return bs
}
