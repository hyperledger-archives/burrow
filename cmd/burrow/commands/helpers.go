package commands

import (
	"fmt"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/genesis"
)

type Output interface {
	Printf(format string, args ...interface{})
	Logf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

func burrowConfigProvider(configFile string) source.ConfigProvider {
	return source.FirstOf(
		// Will fail if file doesn't exist, but still skipped it configFile == ""
		source.File(configFile, false),
		source.Environment(config.DefaultBurrowConfigEnvironmentVariable),
		// Try working directory
		source.File(config.DefaultBurrowConfigTOMLFileName, true),
		source.Default(config.DefaultBurrowConfig()))
}

func genesisDocProvider(genesisFile string, skipNonExistent bool) source.ConfigProvider {
	return source.NewConfigProvider(fmt.Sprintf("genesis file at %s", genesisFile),
		source.ShouldSkipFile(genesisFile, skipNonExistent),
		func(baseConfig interface{}) error {
			conf, ok := baseConfig.(*config.BurrowConfig)
			if !ok {
				return fmt.Errorf("config passed was not BurrowConfig")
			}
			if conf.GenesisDoc != nil {
				return fmt.Errorf("sourcing GenesisDoc from file %v, but GenesisDoc was defined in earlier "+
					"config source, only specify GenesisDoc in one place", genesisFile)
			}
			genesisDoc := new(genesis.GenesisDoc)
			err := source.FromFile(genesisFile, genesisDoc)
			if err != nil {
				return err
			}
			conf.GenesisDoc = genesisDoc
			return nil
		})
}
