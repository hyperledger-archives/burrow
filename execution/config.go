package execution

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/evm"
)

type VMOption string

const (
	DebugOpcodes VMOption = "DebugOpcodes"
	DumpTokens   VMOption = "DumpTokens"
)

type ExecutionConfig struct {
	CallStackMaxDepth        uint64
	DataStackInitialCapacity int
	DataStackMaxDepth        int
	VMOptions                []VMOption `json:",omitempty" toml:",omitempty"`
}

func DefaultExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{
		CallStackMaxDepth:        0, // Unlimited by default
		DataStackInitialCapacity: evm.DataStackInitialCapacity,
		DataStackMaxDepth:        0, // Unlimited by default
	}
}

type ExecutionOption func(*executor)

func VMOptions(vmOptions ...func(*evm.VM)) func(*executor) {
	return func(exe *executor) {
		exe.vmOptions = vmOptions
	}
}

func (ec *ExecutionConfig) ExecutionOptions() ([]ExecutionOption, error) {
	var exeOptions []ExecutionOption
	var vmOptions []func(*evm.VM)
	for _, option := range ec.VMOptions {
		switch option {
		case DebugOpcodes:
			vmOptions = append(vmOptions, evm.DebugOpcodes)
		case DumpTokens:
			vmOptions = append(vmOptions, evm.DumpTokens)
		default:
			return nil, fmt.Errorf("VM option '%s' not recognised", option)
		}
	}
	vmOptions = append(vmOptions, evm.StackOptions(ec.CallStackMaxDepth, ec.DataStackInitialCapacity, ec.DataStackMaxDepth))
	exeOptions = append(exeOptions, VMOptions(vmOptions...))
	return exeOptions, nil
}
