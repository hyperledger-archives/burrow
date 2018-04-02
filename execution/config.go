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
	VMOptions []VMOption `json:",omitempty" toml:",omitempty"`
}

func DefaultExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{}
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
	exeOptions = append(exeOptions, VMOptions(vmOptions...))
	return exeOptions, nil
}
