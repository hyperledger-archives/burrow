package wasm

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/acmstate"
	burrow_binary "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/perlin-network/life/exec"
)

type execContext struct {
	errors.Maybe
	address crypto.Address
	input   []byte
	output  []byte
	state   acmstate.ReaderWriter
}

// Implements ewasm, see https://github.com/ewasm/design

// RunWASM creates a WASM VM, and executes the given WASM contract code
func RunWASM(state acmstate.ReaderWriter, address crypto.Address, createContract bool, wasm, input []byte) (output []byte, cerr error) {
	const errHeader = "ewasm"
	defer func() {
		if r := recover(); r != nil {
			cerr = errors.Codes.ExecutionAborted
		}
	}()

	// WASM
	config := exec.VMConfig{
		DisableFloatingPoint: true,
		MaxMemoryPages:       2,
		DefaultMemoryPages:   2,
	}

	execContext := execContext{
		address: address,
		state:   state,
		input:   input,
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
	case "getCallDataSize":
		return func(vm *exec.VirtualMachine) int64 {
			return int64(len(e.input))
		}

	case "callDataCopy":
		return func(vm *exec.VirtualMachine) int64 {
			destPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			dataOffset := int(uint32(vm.GetCurrentFrame().Locals[1]))
			dataLen := int(uint32(vm.GetCurrentFrame().Locals[2]))

			if dataLen > 0 {
				copy(vm.Memory[destPtr:], e.input[dataOffset:dataOffset+dataLen])
			}

			return 0
		}

	case "storageStore":
		return func(vm *exec.VirtualMachine) int64 {
			keyPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			dataPtr := int(uint32(vm.GetCurrentFrame().Locals[1]))

			key := burrow_binary.Word256{}

			copy(key[:], vm.Memory[keyPtr:keyPtr+32])

			e.Void(e.state.SetStorage(e.address, key, vm.Memory[dataPtr:dataPtr+32]))
			return 0
		}

	case "storageLoad":
		return func(vm *exec.VirtualMachine) int64 {

			keyPtr := int(uint32(vm.GetCurrentFrame().Locals[0]))
			dataPtr := int(uint32(vm.GetCurrentFrame().Locals[1]))

			key := burrow_binary.Word256{}

			copy(key[:], vm.Memory[keyPtr:keyPtr+32])

			val := e.Bytes(e.state.GetStorage(e.address, key))
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

	default:
		panic(fmt.Sprintf("unknown function %s", field))
	}
}

func (e *execContext) ResolveGlobal(module, field string) int64 {
	panic(fmt.Sprintf("global %s module %s not found", field, module))
}
