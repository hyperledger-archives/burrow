package wasm

import (
	"encoding/binary"
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
	state   acmstate.ReaderWriter
}

// In EVM, the code for an account is created by the EVM code itself; the code in the EVM deploy transaction is run,
// and the EVM code returns (via the RETURN) opcode) the code for the contract. In addition, when a new contract is
// created using the "new C()" construct is soldity, the EVM itself passes the code for the new contract.

// The compiler must embed any contract code for smart contracts it wants to create.
// - This does not allow for circular references: contract A creates contract B, contract B creates contract A.
// - This makes it very hard to support other languages; e.g. go or rust have no such bizarre concepts, and would
//   require tricking into supporting this
// - This makes it possible for tricksy contracts that create different contracts at different times. Very hard to
//   support static analysis on these contracts
// - This makes it hard to know ahead-of-time what the code for a contract will be

// Our WASM implementation does not allow for this. The code passed to the deploy transaction, is the contract. Any child contracts must be passed
// during the initial deployment (not implemented yet)

// ABIs
// Our WASM ABI is entirely compatible with Solidity. This means that solidity EVM can call WASM contracts and vice
// versa. However, in the EVM ABI model the difference between constructor calls and function calls are implicit: the
// constructor code path is different from the function call code path, and there is nothing visible in the binary ABI
// encoded arguments telling you that it is a function call or a constructor. Note function calls do have a 4 byte
// prefix but this is not required; without it the fallback function should be called.

// So in our WASM model the smart contract has two entry points: constructor and function.

// ABIs memory space
// In the EVM model, ABIs are passed via the calldata opcodes. This means that the ABI encoded data is not directly
// accessible and has to be copied to smart contract memory. Solidity exposes this via the calldata and memory
// modifiers on variables.

// In our WASM model, the function and constructor WASM fuctions have one argument and return value. The argument is
// where in WASM memory the ABI encoded arguments are and the return value is the offset where the return data can be
// found. At this offset we first find a 32 bit little endian encoded length (since WASM is little endian and we are
// using 32 bit memory model) followed by the bytes themselves.

// Contract Storage
// In the EVM model, contract storage is addressed via 256 bit key and the contents is a 256 bit value. For WASM,
// we've changed the contents to a arbitary byte string. This makes it much easier to store/retrieve smaller values
// (e.g. int64) and dynamic length fields (e.g. strings/arrays).

// Access to contract storage is via WASM externals.
// - set_storage32(uint32 key, uint8* data, uint32 len) // set contract storage
// - get_storage32(uint32 key, uint8* data, uint32 len) // get contract storage (right pad with zeros)

// RunWASM creates a WASM VM, and executes the given WASM contract code
func RunWASM(state acmstate.ReaderWriter, address crypto.Address, createContract bool, wasm, input []byte) (output []byte, cerr error) {
	defer func() {
		if r := recover(); r != nil {
			cerr = errors.Code.ExecutionAborted
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
	}

	vm, err := exec.NewVirtualMachine(wasm, config, &execContext, nil)
	if err != nil {
		return nil, errors.Code.InvalidContract
	}
	if execContext.Error() != nil {
		return nil, execContext.Error()
	}

	// FIXME: Check length
	if len(input) > 0 {
		binary.LittleEndian.PutUint32(vm.Memory[:], uint32(len(input)))
		copy(vm.Memory[4:], input)
	}

	wasmFunc := "function"
	if createContract {
		wasmFunc = "constructor"
	}
	entryID, ok := vm.GetFunctionExport(wasmFunc)
	if !ok {
		return nil, errors.Code.UnresolvedSymbols
	}

	// The 0 argument is the offset where our calldata is stored (if any)
	offset, err := vm.Run(entryID, 0)
	if err != nil {
		return nil, errors.Code.ExecutionAborted
	}

	if offset > 0 {
		// FIXME: Check length
		length := binary.LittleEndian.Uint32(vm.Memory[offset : offset+4])
		output = vm.Memory[offset+4 : offset+4+int64(length)]
	}

	return
}

func (e *execContext) ResolveFunc(module, field string) exec.FunctionImport {
	if module != "env" {
		panic(fmt.Sprintf("unknown module %s", module))
	}

	switch field {
	case "set_storage32":
		return func(vm *exec.VirtualMachine) int64 {
			key := int(uint32(vm.GetCurrentFrame().Locals[0]))
			ptr := int(uint32(vm.GetCurrentFrame().Locals[1]))
			length := int(uint32(vm.GetCurrentFrame().Locals[2]))
			// FIXME: Check length
			e.Void(e.state.SetStorage(e.address, burrow_binary.Int64ToWord256(int64(key)), vm.Memory[ptr:ptr+length]))
			return 0
		}

	case "get_storage32":
		return func(vm *exec.VirtualMachine) int64 {
			key := int(uint32(vm.GetCurrentFrame().Locals[0]))
			ptr := int(uint32(vm.GetCurrentFrame().Locals[1]))
			length := int(uint32(vm.GetCurrentFrame().Locals[2]))
			val := e.Bytes(e.state.GetStorage(e.address, burrow_binary.Int64ToWord256(int64(key))))
			if len(val) < length {
				val = append(val, make([]byte, length-len(val))...)
			}
			// FIXME: Check length
			copy(vm.Memory[ptr:ptr+length], val)
			return 0
		}
	default:
		panic(fmt.Sprintf("unknown function %s", field))
	}
}

func (e *execContext) ResolveGlobal(module, field string) int64 {
	panic(fmt.Sprintf("global %s module %s not found", field, module))
}
