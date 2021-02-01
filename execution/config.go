package execution

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/engine"

	"github.com/hyperledger/burrow/execution/evm"
)

type VMOption string

const (
	DebugOpcodes VMOption = "DebugOpcodes"
	DumpTokens   VMOption = "DumpTokens"
)

type ExecutionConfig struct {
	// This parameter scales the default Tendermint timeouts. A value of 1 gives the Tendermint defaults designed to
	// work for 100 node + public network. Smaller networks should be able to sustain lower values.
	// When running in no-consensus mode (Tendermint.Enabled = false) this scales the block duration with 1.0 meaning 1 second
	// and 0 meaning commit immediately
	TimeoutFactor            float64
	CallStackMaxDepth        uint64
	DataStackInitialCapacity uint64
	DataStackMaxDepth        uint64
	VMOptions                []VMOption `json:",omitempty" toml:",omitempty"`
}

func DefaultExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{
		CallStackMaxDepth:        0, // Unlimited by default
		DataStackInitialCapacity: evm.DataStackInitialCapacity,
		DataStackMaxDepth:        0, // Unlimited by default
		TimeoutFactor:            0.33,
	}
}

type Option func(*executor)

func VMOptions(vmOptions evm.Options) func(*executor) {
	return func(exe *executor) {
		exe.vmOptions = vmOptions
	}
}

func (ec *ExecutionConfig) ExecutionOptions() ([]Option, error) {
	var exeOptions []Option
	vmOptions := evm.Options{
		MemoryProvider:           engine.DefaultDynamicMemoryProvider,
		CallStackMaxDepth:        ec.CallStackMaxDepth,
		DataStackInitialCapacity: ec.DataStackInitialCapacity,
		DataStackMaxDepth:        ec.DataStackMaxDepth,
	}
	for _, option := range ec.VMOptions {
		switch option {
		case DebugOpcodes:
			vmOptions.DebugOpcodes = true
		case DumpTokens:
			vmOptions.DumpTokens = true
		default:
			return nil, fmt.Errorf("VM option '%s' not recognised", option)
		}
	}
	exeOptions = append(exeOptions, VMOptions(vmOptions))
	return exeOptions, nil
}
