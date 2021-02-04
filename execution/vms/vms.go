package vms

import (
	"github.com/hyperledger/burrow/execution/defaults"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/wasm"
)

type VirtualMachines struct {
	*evm.EVM
	*wasm.WVM
}

func NewConnectedVirtualMachines(options engine.Options) *VirtualMachines {
	options = defaults.CompleteOptions(options)
	evm := evm.New(options)
	wvm := wasm.New(options)
	// Allow the virtual machines to call each other
	engine.Connect(evm, wvm)
	return &VirtualMachines{
		EVM: evm,
		WVM: wvm,
	}
}
