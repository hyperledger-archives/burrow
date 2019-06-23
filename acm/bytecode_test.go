package acm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBytecode_MarshalJSON(t *testing.T) {
	bytecode := Bytecode{
		73, 234, 48, 252, 174,
		115, 27, 222, 54, 116,
		47, 133, 144, 21, 73,
		245, 21, 234, 26, 50,
	}

	bs, err := json.Marshal(bytecode)
	assert.NoError(t, err)

	bytecodeOut := new(Bytecode)
	err = json.Unmarshal(bs, bytecodeOut)
	require.NoError(t, err)

	assert.Equal(t, bytecode, *bytecodeOut)
}

func TestBytecode_MarshalText(t *testing.T) {
	bytecode := Bytecode{
		73, 234, 48, 252, 174,
		115, 27, 222, 54, 116,
		47, 133, 144, 21, 73,
		245, 21, 234, 26, 50,
	}

	bs, err := bytecode.MarshalText()
	assert.NoError(t, err)

	bytecodeOut := new(Bytecode)
	err = bytecodeOut.UnmarshalText(bs)
	require.NoError(t, err)

	assert.Equal(t, bytecode, *bytecodeOut)
}

func TestBytecode_Tokens(t *testing.T) {
	/*
			pragma solidity ^0.5.4;

			contract SimpleStorage {
		                function get() public constant returns (address) {
		        	        return msg.sender;
		    	        }
			}
	*/

	// This bytecode is compiled from Solidity contract above the remix compiler with 0.4.0
	codeHex := "606060405260808060106000396000f360606040526000357c0100000000000000000000000000000000000000000000000" +
		"000000000900480636d4ce63c146039576035565b6002565b34600257604860048050506074565b604051808273fffffffffffffff" +
		"fffffffffffffffffffffffff16815260200191505060405180910390f35b6000339050607d565b9056"
	bytecode, err := BytecodeFromHex(codeHex)
	require.NoError(t, err)
	tokens, err := bytecode.Tokens()
	require.NoError(t, err)
	// With added leading zero in hex where needed
	remixOpcodes := "PUSH1 0x60 PUSH1 0x40 MSTORE PUSH1 0x80 DUP1 PUSH1 0x10 PUSH1 0x00 CODECOPY PUSH1 0x00 RETURN " +
		"PUSH1 0x60 PUSH1 0x40 MSTORE PUSH1 0x00 CALLDATALOAD " +
		"PUSH29 0x0100000000000000000000000000000000000000000000000000000000 SWAP1 DIV DUP1 PUSH4 0x6D4CE63C EQ " +
		"PUSH1 0x39 JUMPI PUSH1 0x35 JUMP JUMPDEST PUSH1 0x02 JUMP JUMPDEST CALLVALUE PUSH1 0x02 JUMPI " +
		"PUSH1 0x48 PUSH1 0x04 DUP1 POP POP PUSH1 0x74 JUMP JUMPDEST PUSH1 0x40 MLOAD DUP1 DUP3 " +
		"PUSH20 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF AND DUP2 MSTORE PUSH1 0x20 ADD " +
		"SWAP2 POP POP PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 RETURN JUMPDEST PUSH1 0x00 " +
		"CALLER SWAP1 POP PUSH1 0x7D JUMP JUMPDEST SWAP1 JUMP"
	assert.Equal(t, remixOpcodes, strings.Join(tokens, " "))

	// Test empty bytecode
	tokens, err = Bytecode(nil).Tokens()
	require.NoError(t, err)
	assert.Equal(t, []string{}, tokens)

	_, err = Bytecode(bc.MustSplice(asm.PUSH3, 1, 2)).Tokens()
	assert.Error(t, err, "not enough bytes to push")
}
