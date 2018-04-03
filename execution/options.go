package execution

import "github.com/hyperledger/burrow/execution/evm"

type ExecutionOption func(*executor)

func VMOptions(vmOptions ...func(*evm.VM)) func(*executor) {
	return func(exe *executor) {
		exe.vmOptions = vmOptions
	}
}
