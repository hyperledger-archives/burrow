package bc

import (
	"github.com/hyperledger/burrow/word"
	"github.com/hyperledger/burrow/account"
	"fmt"
	"github.com/hyperledger/burrow/execution/evm/asm"
)

// Convenience function to allow us to mix bytes, ints, and OpCodes that
// represent bytes in an EVM assembly code to make assembly more readable.
// Also allows us to splice together assembly
// fragments because any []byte arguments are flattened in the result.
func Splice(bytelikes ...interface{}) []byte {
	bytes := make([]byte, len(bytelikes))
	for i, bytelike := range bytelikes {
		switch b := bytelike.(type) {
		case byte:
			bytes[i] = b
		case asm.OpCode:
			bytes[i] = byte(b)
		case int:
			bytes[i] = byte(b)
			if int(bytes[i]) != b {
				panic(fmt.Sprintf("The int %v does not fit inside a byte", b))
			}
		case int64:
			bytes[i] = byte(b)
			if int64(bytes[i]) != b {
				panic(fmt.Sprintf("The int64 %v does not fit inside a byte", b))
			}
		case uint64:
			bytes[i] = byte(b)
			if uint64(bytes[i]) != b {
				panic(fmt.Sprintf("The uint64 %v does not fit inside a byte", b))
			}
		case word.Word256:
			return Concat(bytes[:i], b[:], Splice(bytelikes[i+1:]...))
		case word.Word160:
			return Concat(bytes[:i], b[:], Splice(bytelikes[i+1:]...))
		case account.Address:
			return Concat(bytes[:i], b[:], Splice(bytelikes[i+1:]...))
		case []byte:
			// splice
			return Concat(bytes[:i], b, Splice(bytelikes[i+1:]...))
		default:
			panic(fmt.Errorf("could not convert %s to a byte or sequence of bytes", bytelike))
		}
	}
	return bytes
}

func Concat(bss ...[]byte) []byte {
	offset := 0
	for _, bs := range bss {
		offset += len(bs)
	}
	bytes := make([]byte, offset)
	offset = 0
	for _, bs := range bss {
		for i, b := range bs {
			bytes[offset+i] = b
		}
		offset += len(bs)
	}
	return bytes
}
