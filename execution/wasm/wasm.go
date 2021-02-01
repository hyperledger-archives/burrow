package wasm

import (
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/native"
	"github.com/hyperledger/burrow/logging"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	lifeExec "github.com/perlin-network/life/exec"
)

// Implements ewasm, see https://github.com/ewasm/design
// WASM
var DefaultVMConfig = lifeExec.VMConfig{
	DisableFloatingPoint: true,
	MaxMemoryPages:       16,
	DefaultMemoryPages:   16,
}

type WVM struct {
	engine.Externals
	options  Options
	vmConfig lifeExec.VMConfig
}

type Options struct {
	Natives           *native.Natives
	CallStackMaxDepth uint64
	Logger            *logging.Logger
}

func New(options Options) *WVM {
	if options.Natives == nil {
		options.Natives = native.MustDefaultNatives()
	}
	if options.Logger == nil {
		options.Logger = logging.NewNoopLogger()
	}
	return &WVM{
		options:  options,
		vmConfig: DefaultVMConfig,
	}
}

func Default() *WVM {
	return New(Options{})
}

// RunWASM creates a WASM VM, and executes the given WASM contract code
func (vm *WVM) Execute(st acmstate.ReaderWriter, blockchain engine.Blockchain, eventSink exec.EventSink,
	params engine.CallParams, code []byte) (output []byte, cerr error) {
	defer func() {
		if r := recover(); r != nil {
			cerr = errors.Codes.ExecutionAborted
		}
	}()

	st = native.NewState(vm.options.Natives, st)

	state := engine.State{
		CallFrame:  engine.NewCallFrame(st).WithMaxCallStackDepth(vm.options.CallStackMaxDepth),
		Blockchain: blockchain,
		EventSink:  eventSink,
	}

	output, err := vm.Contract(code).Call(state, params)

	if err == nil {
		// Only sync back when there was no exception
		err = state.CallFrame.Sync()
	}
	// Always return output - we may have a reverted exception for which the return is meaningful
	return output, err
}

func (vm *WVM) Dispatch(acc *acm.Account) engine.Callable {
	callable := vm.Externals.Dispatch(acc)
	if callable != nil {
		return callable
	}
	return vm.Contract(acc.WASMCode)
}

func (vm *WVM) Contract(code []byte) *Contract {
	return &Contract{
		vm:   vm,
		code: code,
	}
}
