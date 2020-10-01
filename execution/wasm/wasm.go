package wasm

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/hyperledger/burrow/acm/acmstate"
	burrow_binary "github.com/hyperledger/burrow/binary"
	crypto "github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/perlin-network/life/exec"
)

type execContext struct {
	errors.Maybe
	code       []byte
	output     []byte
	returnData []byte
	params     engine.CallParams
	state      acmstate.ReaderWriter
}

// Implements ewasm, see https://github.com/ewasm/design

// RunWASM creates a WASM VM, and executes the given WASM contract code
func RunWASM(state acmstate.ReaderWriter, params engine.CallParams, wasm []byte) (output []byte, cerr error) {
	const errHeader = "ewasm"
	defer func() {
		if r := recover(); r != nil {
			cerr = errors.Codes.ExecutionAborted
		}
	}()

	// WASM
	config := exec.VMConfig{
		DisableFloatingPoint: true,
		MaxMemoryPages:       16,
		DefaultMemoryPages:   16,
	}

	execContext := execContext{
		params: params,
		code:   wasm,
		state:  state,
	}

	// panics in ResolveFunc() will be recovered for us, no need for our own
	vm, err := exec.NewVirtualMachine(wasm, config, &execContext, nil)
	if err != nil {
		return nil, errors.Errorf(errors.Codes.InvalidContract, "%s: %v", errHeader, err)
	}
	if execContext.Error() != nil {
		return nil, execContext.Error()
	}

	entryID, ok := vm.GetFunctionExport("main")
	if !ok {
		return nil, errors.Codes.UnresolvedSymbols
	}

	_, err = vm.Run(entryID)
	if err != nil && errors.GetCode(err) != errors.Codes.None {
		return nil, errors.Errorf(errors.Codes.ExecutionAborted, "%s: %v", errHeader, err)
	}

	return execContext.output, nil
}

func (e *execContext) ResolveFunc(module, field string) exec.FunctionImport {
	if module != "ethereum" {
		panic(fmt.Sprintf("unknown module %s", module))
	}

	switch field {
	case "call":
		return func(vm *exec.VirtualMachine) int64 {
			// gas := int(uint64(vm.GetCurrentFrame().Locals[0]))
			addressPtr := int(uint32(vm.GetCurrentFrame().Locals[1]))
			// valuePtr := int(uint32(vm.GetCurrentFrame().Locals[2]))
			dataPtr := int(uint32(vm.GetCurrentFrame().Locals[3]))
			dataLen := int(uint32(vm.GetCurrentFrame().Locals[4]))

			// fixed support for system contract of keccak256
			address := make([]byte, 20)

			copy(address[:], vm.Memory[addressPtr:addressPtr+crypto.AddressLength])

			if bytes.Equal(address, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}) {
				copy(e.returnData[0:32], crypto.SHA256(vm.Memory[dataPtr:dataPtr+dataLen]))
			} else if bytes.Equal(address, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03}) {
				copy(e.returnData[0:32], crypto.RIPEMD160(vm.Memory[dataPtr:dataPtr+dataLen]))
			} else if bytes.Equal(address, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09}) {
				copy(e.returnData[0:32], crypto.Keccak256(vm.Memory[dataPtr:dataPtr+dataLen]))
			} else {
				panic(errors.Codes.InvalidAddress)
			}

			return 1
		}

	case "getCallDataSize":
		return func(vm *exec.VirtualMachine) int64 {
			return int64(len(e.params.Input))
		}

	case "callDataCopy":
		return func(vm *exec.VirtualMachine) int64 {
			destPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			dataOffset := int(uint32(vm.GetCurrentFrame().Locals[1]))
			dataLen := int(uint32(vm.GetCurrentFrame().Locals[2]))

			if dataLen > 0 {
				copy(vm.Memory[destPtr:], e.params.Input[dataOffset:dataOffset+dataLen])
			}

			return 0
		}

	case "getReturnDataSize":
		return func(vm *exec.VirtualMachine) int64 {
			return int64(len(e.returnData))
		}

	case "returnDataCopy":
		return func(vm *exec.VirtualMachine) int64 {
			destPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			dataOffset := int(uint32(vm.GetCurrentFrame().Locals[1]))
			dataLen := int(uint32(vm.GetCurrentFrame().Locals[2]))

			if dataLen > 0 {
				copy(vm.Memory[destPtr:], e.returnData[dataOffset:dataOffset+dataLen])
			}

			return 0
		}

	case "getCodeSize":
		return func(vm *exec.VirtualMachine) int64 {
			return int64(len(e.code))
		}

	case "codeCopy":
		return func(vm *exec.VirtualMachine) int64 {
			destPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			dataOffset := int(uint32(vm.GetCurrentFrame().Locals[1]))
			dataLen := int(uint32(vm.GetCurrentFrame().Locals[2]))

			if dataLen > 0 {
				copy(vm.Memory[destPtr:], e.code[dataOffset:dataOffset+dataLen])
			}

			return 0
		}

	case "storageStore":
		return func(vm *exec.VirtualMachine) int64 {
			keyPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			dataPtr := int(uint32(vm.GetCurrentFrame().Locals[1]))

			key := burrow_binary.Word256{}

			copy(key[:], vm.Memory[keyPtr:keyPtr+32])

			e.Void(e.state.SetStorage(e.params.Callee, key, vm.Memory[dataPtr:dataPtr+32]))
			return 0
		}

	case "storageLoad":
		return func(vm *exec.VirtualMachine) int64 {

			keyPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			dataPtr := int(uint32(vm.GetCurrentFrame().Locals[1]))

			key := burrow_binary.Word256{}

			copy(key[:], vm.Memory[keyPtr:keyPtr+32])

			val := e.Bytes(e.state.GetStorage(e.params.Callee, key))
			copy(vm.Memory[dataPtr:], val)

			return 0
		}

	case "finish":
		return func(vm *exec.VirtualMachine) int64 {
			dataPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			dataLen := int(uint32(vm.GetCurrentFrame().Locals[1]))

			e.output = vm.Memory[dataPtr : dataPtr+dataLen]

			panic(errors.Codes.None)
		}

	case "revert":
		return func(vm *exec.VirtualMachine) int64 {

			dataPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			dataLen := int(uint32(vm.GetCurrentFrame().Locals[1]))

			e.output = vm.Memory[dataPtr : dataPtr+dataLen]

			panic(errors.Codes.ExecutionReverted)
		}

	case "getAddress":
		return func(vm *exec.VirtualMachine) int64 {
			addressPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))

			copy(vm.Memory[addressPtr:], e.params.Callee.Bytes())

			return 0
		}

	case "getCallValue":
		return func(vm *exec.VirtualMachine) int64 {

			valuePtr := int(uint32(vm.GetCurrentFrame().Locals[0]))

			// ewasm value is little endian 128 bit value
			bs := make([]byte, 16)
			binary.LittleEndian.PutUint64(bs, e.params.Value)

			copy(vm.Memory[valuePtr:], bs)

			return 0
		}

	case "getExternalBalance":
		return func(vm *exec.VirtualMachine) int64 {
			addressPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			balancePtr := int(uint32(vm.GetCurrentFrame().Locals[1]))

			address := crypto.Address{}

			copy(address[:], vm.Memory[addressPtr:addressPtr+crypto.AddressLength])
			acc, err := e.state.GetAccount(address)
			if err != nil {
				panic(errors.Codes.InvalidAddress)
			}

			// ewasm value is little endian 128 bit value
			bs := make([]byte, 16)
			binary.LittleEndian.PutUint64(bs, acc.Balance)

			copy(vm.Memory[balancePtr:], bs)

			return 0
		}

	default:
		panic(fmt.Sprintf("unknown function %s", field))
	}
}

func (e *execContext) ResolveGlobal(module, field string) int64 {
	panic(fmt.Sprintf("global %s module %s not found", field, module))
}
