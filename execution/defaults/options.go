package defaults

import (
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/native"
	"github.com/hyperledger/burrow/logging"
)

func CompleteOptions(options engine.Options) engine.Options {
	// Set defaults
	if options.MemoryProvider == nil {
		options.MemoryProvider = engine.DefaultDynamicMemoryProvider
	}
	if options.Logger == nil {
		options.Logger = logging.NewNoopLogger()
	}
	if options.Natives == nil {
		options.Natives = native.MustDefaultNatives()
	}
	return options
}
