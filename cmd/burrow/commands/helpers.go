package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/genesis"
	logging_config "github.com/hyperledger/burrow/logging/logconfig"
)

type Output interface {
	Printf(format string, args ...interface{})
	Logf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

func obtainBurrowConfig(configFile, genesisDocFile string) (*config.BurrowConfig, error) {
	// We need to reflect on whether this obscures where values are coming from
	conf := config.DefaultBurrowConfig()
	// We treat logging a little differently in that if anything is set for logging we will not
	// set default outputs
	conf.Logging = nil
	err := source.EachOf(
		burrowConfigProvider(configFile),
		source.FirstOf(
			genesisDocProvider(genesisDocFile, false),
			// Try working directory
			genesisDocProvider(config.DefaultGenesisDocJSONFileName, true)),
	).Apply(conf)
	if err != nil {
		return nil, err
	}
	// If no logging config was provided use the default
	if conf.Logging == nil {
		conf.Logging = logging_config.DefaultNodeLoggingConfig()
	}
	return conf, nil
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

func parseRange(rangeString string) (start int64, end int64, err error) {
	start = 0
	end = -1

	if rangeString == "" {
		return
	}

	bounds := strings.Split(rangeString, ":")
	if len(bounds) == 1 {
		startString := bounds[0]
		start, err = strconv.ParseInt(startString, 10, 64)
		return
	}
	if len(bounds) == 2 {
		if bounds[0] != "" {
			start, err = strconv.ParseInt(bounds[0], 10, 64)
			if err != nil {
				return
			}
		}
		if bounds[1] != "" {
			end, err = strconv.ParseInt(bounds[1], 10, 64)
		}
		return
	}
	return 0, 0, fmt.Errorf("could not parse range from %s", rangeString)
}
