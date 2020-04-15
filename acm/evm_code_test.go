package acm

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	"github.com/stretchr/testify/assert"
	"github.com/tmthrgd/go-bitset"
)

func TestOpcodeBitset(t *testing.T) {
	tests := []struct {
		name   string
		code   []byte
		bitset bitset.Bitset
	}{
		{
			name:   "Only one real JUMPDEST",
			code:   bc.MustSplice(asm.PUSH2, 1, asm.JUMPDEST, asm.ADD, asm.JUMPDEST),
			bitset: mkBitset("10011"),
		},
		{
			name:   "Two JUMPDESTs",
			code:   bc.MustSplice(asm.PUSH1, 1, asm.JUMPDEST, asm.ADD, asm.JUMPDEST),
			bitset: mkBitset("10111"),
		},
		{
			name:   "One PUSH",
			code:   bc.MustSplice(asm.PUSH4, asm.JUMPDEST, asm.ADD, asm.JUMPDEST, asm.PUSH32, asm.BALANCE),
			bitset: mkBitset("100001"),
		},
		{
			name:   "Two PUSHes",
			code:   bc.MustSplice(asm.PUSH3, asm.JUMPDEST, asm.ADD, asm.JUMPDEST, asm.PUSH32, asm.BALANCE),
			bitset: mkBitset("100010"),
		},
		{
			name:   "Three PUSHes",
			code:   bc.MustSplice(asm.PUSH3, asm.JUMPDEST, asm.ADD, asm.PUSH2, asm.PUSH32, asm.BALANCE),
			bitset: mkBitset("100010"),
		},
		{
			name:   "No PUSH",
			code:   bc.MustSplice(asm.JUMPDEST, asm.ADD, asm.BALANCE),
			bitset: mkBitset("111"),
		},
		{
			name:   "End PUSH",
			code:   bc.MustSplice(asm.JUMPDEST, asm.ADD, asm.PUSH6),
			bitset: mkBitset("111"),
		},
		{
			name:   "Middle PUSH",
			code:   bc.MustSplice(asm.JUMPDEST, asm.PUSH2, asm.PUSH1, asm.PUSH2, asm.BLOCKHASH),
			bitset: mkBitset("11001"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EVMOpcodeBitset(tt.code); !reflect.DeepEqual(got, tt.bitset) {
				t.Errorf("EVMOpcodeBitset() = %v, want %v", got, tt.bitset)
			}
		})
	}
}

func TestEVMCode_IsOpcodeAt(t *testing.T) {
	code := MustEVMCodeFrom(asm.PUSH2, 2, 3)

	assert.True(t, code.IsOpcode(0))
	assert.False(t, code.IsOpcode(1))
	assert.False(t, code.IsOpcode(2))
	assert.False(t, code.IsOpcode(3))
}

func mkBitset(binaryString string) bitset.Bitset {
	length := uint(len(binaryString))
	bs := bitset.New(length)
	for i := uint(0); i < length; i++ {
		switch binaryString[i] {
		case '1':
			bs.Set(i)
		case '0':
		case ' ':
			i++
		default:
			panic(fmt.Errorf("mkBitset() expects a string containing only 1s, 0s, and spaces"))
		}
	}
	return bs
}
