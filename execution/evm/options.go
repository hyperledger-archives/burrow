package evm

func MemoryProvider(memoryProvider func() Memory) func(*VM) {
	return func(vm *VM) {
		vm.memoryProvider = memoryProvider
	}
}

func DebugOpcodes(vm *VM) {
	vm.debugOpcodes = true
}

func DumpTokens(vm *VM) {
	vm.dumpTokens = true
}
