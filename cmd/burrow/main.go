package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/jawher/mow.cli"
)

func main() {
	burrow := cli.App("burrow", "The EVM smart contract machine with Tendermint consensus")

	genesisOpt := burrow.StringOpt("g genesis", "",
		"Use the specified genesis JSON file rather than a key in the main config, use - to read from STDIN")

	configOpt := burrow.StringOpt("c config", "",
		"Use the a specified burrow config TOML file")

	burrow.Spec = "[--config=<config file>] [--genesis=<genesis json file>]"

	burrow.Action = func() {
		conf := config.DefaultBurrowConfig()
		err := burrowConfigProvider(*configOpt, *genesisOpt).Apply(conf)
		if err != nil {
			fatalf("could not obtain config: %v", err)
		}

		kern, err := conf.Kernel()
		if err != nil {
			fatalf("could not create Burrow kernel: %v", err)
		}

		err = kern.Boot()
		if err != nil {
			fatalf("could not boot Burrow kernel: %v", err)
		}
		kern.WaitForShutdown()
	}

	burrow.Command("spec",
		"Build a GenesisSpec that acts as a template for a GenesisDoc and the configure command",
		func(cmd *cli.Cmd) {
			tomlOpt := cmd.BoolOpt("t toml", false, "Emit GenesisSpec as TOML rather than the "+
				"default JSON")

			participantsOpt := cmd.IntOpt("p participant-accounts", 1, "Number of preset Participant type accounts")

			fullOpt := cmd.IntOpt("f full-accounts", 1, "Number of preset Full type accounts")

			cmd.Spec = "[--participant-accounts] [--full-accounts] [--toml]"

			cmd.Action = func() {
				specs := make([]spec.GenesisSpec, 0, *participantsOpt+*fullOpt)
				for i := 0; i < *participantsOpt; i++ {
					specs = append(specs, spec.ParticipantAccount(i))
				}
				for i := 0; i < *fullOpt; i++ {
					specs = append(specs, spec.FullAccount(i))
				}
				genesisSpec := spec.MergeGenesisSpecs(specs...)
				if *tomlOpt {
					os.Stdout.WriteString(source.TOMLString(genesisSpec))
				} else {
					os.Stdout.WriteString(source.JSONString(genesisSpec))
				}
			}
		})

	burrow.Command("configure",
		"Create Burrow configuration by consuming a GenesisDoc or GenesisSpec, creating keys, and emitting the config",
		func(cmd *cli.Cmd) {
			genesisSpecOpt := cmd.StringOpt("s genesis-spec", "",
				"A GenesisSpec to use as a tmeplate for a GenesisDoc that will be created along with keys")

			tomlInOpt := cmd.BoolOpt("t toml-in", false, "Consume GenesisSpec/GenesisDoc as TOML "+
				"rather than the JSON default")

			jsonOutOpt := cmd.BoolOpt("j json-out", false, "Emit config in JSON rather than TOML "+
				"suitable for further processing or forming a separate genesis.json GenesisDoc")

			genesisDocOpt := cmd.StringOpt("g genesis-doc", "", "GenesisDoc JSON to embed in config")

			validatorIndexOpt := cmd.IntOpt("v validator-index", -1,
				"Validator index (in validators list - GenesisSpec or GenesisDoc) from which to set ValidatorAddress")

			cmd.Spec = "[--genesis-spec=<GenesisSpec file> | --genesis-doc=<GenesisDoc file>] " +
				"[--validator-index=<index>] [--toml-in] [--json-out]"

			cmd.Action = func() {
				conf := config.DefaultBurrowConfig()
				err := burrowConfigProvider(*configOpt, *genesisOpt).Apply(conf)
				if err != nil {
					fatalf("could not obtain config: %v", err)
				}
				if *genesisSpecOpt != "" {
					genesisSpec := new(spec.GenesisSpec)
					err := fromFile(*genesisSpecOpt, *tomlInOpt, genesisSpec)
					if err != nil {
						fatalf("could not read GenesisSpec: %v", err)
					}
					conf.GenesisDoc, err = conf.RealiseGenesisSpec(genesisSpec)
					if err != nil {
						fatalf("could not realise GenesisSpec: %v", err)
					}
				} else if *genesisDocOpt != "" {
					genesisDoc := new(genesis.GenesisDoc)
					err := fromFile(*genesisSpecOpt, *tomlInOpt, genesisDoc)
					if err != nil {
						fatalf("could not read GenesisSpec: %v", err)
					}
					conf.GenesisDoc = genesisDoc
				}
				if *validatorIndexOpt > -1 {
					if conf.GenesisDoc == nil {
						fatalf("Unable to set ValidatorAddress from provided validator-index since no " +
							"GenesisDoc/GenesisSpec provided.")
					}
					if len(conf.GenesisDoc.Validators) < *validatorIndexOpt {
						fatalf("validator-index of %v given but only %v validators specified in GenesisDoc",
							*validatorIndexOpt, len(conf.GenesisDoc.Validators))
					}
					conf.ValidatorAddress = &conf.GenesisDoc.Validators[*validatorIndexOpt].Address
				}
				if *jsonOutOpt {
					os.Stdout.WriteString(conf.JSONString())
				} else {
					os.Stdout.WriteString(conf.TOMLString())
				}
			}
		})

	burrow.Run(os.Args)
}

// Print informational output to Stderr
func printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func burrowConfigProvider(configFile, genesisFile string) source.ConfigProvider {
	return source.EachOf(
		source.FirstOf(
			// Will fail if file doesn't exist, but still skipped it configFile == ""
			source.TOMLFile(configFile, false),
			source.Environment(config.DefaultBurrowConfigJSONEnvironmentVariable),
			// Try working directory
			source.TOMLFile(config.DefaultBurrowConfigTOMLFileName, true),
			source.Default(config.DefaultBurrowConfig())),
		source.FirstOf(
			genesisDocProvider(genesisFile, false),
			// Try working directory
			genesisDocProvider(config.DefaultGenesisDocJSONFileName, true)),
	)
}

func genesisDocProvider(genesisFile string, skipNonExistent bool) source.ConfigProvider {
	return source.NewConfigProvider(fmt.Sprintf("genesis file at %s", genesisFile),
		source.ShouldSkipFile(genesisFile, skipNonExistent),
		func(baseConfig interface{}) error {
			conf, ok := baseConfig.(config.BurrowConfig)
			if !ok {
				return fmt.Errorf("config passed was not BurrowConfig")
			}
			if conf.GenesisDoc != nil {
				return fmt.Errorf("sourcing GenesisDoc from file, but GenesisDoc was defined earlier " +
					"in config cascade, only specify GenesisDoc in one place")
			}
			genesisDoc := new(genesis.GenesisDoc)
			err := source.FromJSONFile(genesisFile, genesisDoc)
			if err != nil {
				return err
			}
			conf.GenesisDoc = genesisDoc
			return nil
		})
}

func fromFile(file string, toml bool, conf interface{}) (err error) {
	if toml {
		err = source.FromTOMLFile(file, conf)
		if err != nil {
			fatalf("could not read GenesisSpec: %v", err)
		}
	} else {
		err = source.FromJSONFile(file, conf)
		if err != nil {
			fatalf("could not read GenesisSpec: %v", err)
		}
	}
	return
}
