package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	logging_config "github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/config/presets"
	"github.com/hyperledger/burrow/project"
	"github.com/jawher/mow.cli"
)

func main() {
	burrow := cli.App("burrow", "The EVM smart contract machine with Tendermint consensus")

	genesisOpt := burrow.StringOpt("g genesis", "",
		"Use the specified genesis JSON file rather than a key in the main config, use - to read from STDIN")

	configOpt := burrow.StringOpt("c config", "",
		"Use the a specified burrow config TOML file")

	versionOpt := burrow.BoolOpt("v version", false, "Print the Burrow version")

	burrow.Spec = "[--config=<config file>] [--genesis=<genesis json file>] [--version]"

	burrow.Action = func() {
		if *versionOpt {
			fmt.Println(project.History.CurrentVersion().String())
			os.Exit(0)
		}
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
		// We need to reflect on whether this obscures where values are coming from
		conf := config.DefaultBurrowConfig()
		// We treat logging a little differently in that if anything is set for logging we will not
		// set default outputs
		conf.Logging = nil
		err := source.EachOf(
			burrowConfigProvider(*configOpt),
			source.FirstOf(
				genesisDocProvider(*genesisOpt, false),
				// Try working directory
				genesisDocProvider(config.DefaultGenesisDocJSONFileName, true)),
		).Apply(conf)
		// If no logging config was provided use the default
		if conf.Logging == nil {
			conf.Logging = logging_config.DefaultNodeLoggingConfig()
		}
		if err != nil {
			fatalf("could not obtain config: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		kern, err := conf.Kernel(ctx)
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

			fullOpt := cmd.IntOpt("f full-accounts", 1, "Number of preset Full type accounts")
			validatorOpt := cmd.IntOpt("v validator-accounts", 0, "Number of preset Validator type accounts")
			rootOpt := cmd.IntOpt("r root-accounts", 0, "Number of preset Root type accounts")
			developerOpt := cmd.IntOpt("d developer-accounts", 0, "Number of preset Developer type accounts")
			participantsOpt := cmd.IntOpt("p participant-accounts", 1, "Number of preset Participant type accounts")

			cmd.Spec = "[--full-accounts] [--validator-accounts] [--root-accounts] [--developer-accounts] [--participant-accounts] [--toml]"

			cmd.Action = func() {
				specs := make([]spec.GenesisSpec, 0, *participantsOpt+*fullOpt)
				for i := 0; i < *fullOpt; i++ {
					specs = append(specs, spec.FullAccount(i))
				}
				for i := 0; i < *validatorOpt; i++ {
					specs = append(specs, spec.ValidatorAccount(i))
				}
				for i := 0; i < *rootOpt; i++ {
					specs = append(specs, spec.RootAccount(i))
				}
				for i := 0; i < *developerOpt; i++ {
					specs = append(specs, spec.DeveloperAccount(i))
				}
				for i := 0; i < *participantsOpt; i++ {
					specs = append(specs, spec.ParticipantAccount(i))
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
				"A GenesisSpec to use as a template for a GenesisDoc that will be created along with keys")

			tomlInOpt := cmd.BoolOpt("t toml-in", false, "Consume GenesisSpec/GenesisDoc as TOML "+
				"rather than the JSON default")

			keysUrlOpt := cmd.StringOpt("k keys-url", "", fmt.Sprintf("Provide keys URL, default: %s",
				keys.DefaultKeysConfig().URL))

			jsonOutOpt := cmd.BoolOpt("j json-out", false, "Emit config in JSON rather than TOML "+
				"suitable for further processing or forming a separate genesis.json GenesisDoc")

			genesisDocOpt := cmd.StringOpt("g genesis-doc", "", "GenesisDoc JSON to embed in config")

			validatorIndexOpt := cmd.IntOpt("v validator-index", -1,
				"Validator index (in validators list - GenesisSpec or GenesisDoc) from which to set ValidatorAddress")

			loggingOpt := cmd.StringOpt("l logging", "",
				"Comma separated list of logging instructions which form a 'program' which is a depth-first "+
					"pre-order of instructions that will build the root logging sink. See 'burrow help' for more information.")

			describeLoggingOpt := cmd.BoolOpt("describe-logging", false,
				"Print an exhaustive list of logging instructions available with the --logging option")

			cmd.Spec = "[--keys-url=<keys URL>] [--genesis-spec=<GenesisSpec file> | --genesis-doc=<GenesisDoc file>] " +
				"[--validator-index=<index>] [--toml-in] [--json-out] [--logging=<logging program>] [--describe-logging]"

			cmd.Action = func() {
				conf := config.DefaultBurrowConfig()

				if *configOpt != "" {
					// If explicitly given a config file use it as a base:
					err := source.FromTOMLFile(*configOpt, conf)
					if err != nil {
						fatalf("could not read base config file (as TOML): %v", err)
					}
				}

				if *describeLoggingOpt {
					fmt.Printf("Usage:\n  burrow configure -l INSTRUCTION[,...]\n\nBuilds a logging " +
						"configuration by constructing a tree of logging sinks assembled from preset instructions " +
						"that generate the tree while traversing it.\n\nLogging Instructions:\n")
					for _, instruction := range presets.Instructons() {
						fmt.Printf("  %-15s\t%s\n", instruction.Name(), instruction.Description())
					}
					fmt.Printf("\nExample Usage:\n  burrow configure -l include-any,info,stderr\n")
					return
				}

				if *keysUrlOpt != "" {
					conf.Keys.URL = *keysUrlOpt
				}

				if *genesisSpecOpt != "" {
					genesisSpec := new(spec.GenesisSpec)
					err := fromFile(*genesisSpecOpt, *tomlInOpt, genesisSpec)
					if err != nil {
						fatalf("could not read GenesisSpec: %v", err)
					}
					keyClient := keys.NewKeyClient(conf.Keys.URL, logging.NewNoopLogger())
					conf.GenesisDoc, err = genesisSpec.GenesisDoc(keyClient)
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
				} else if conf.GenesisDoc != nil && len(conf.GenesisDoc.Validators) > 0 {
					// Pick first validator otherwise - might want to change this when we support non-validating node
					conf.ValidatorAddress = &conf.GenesisDoc.Validators[0].Address
				}

				if *loggingOpt != "" {
					ops := strings.Split(*loggingOpt, ",")
					sinkConfig, err := presets.BuildSinkConfig(ops...)
					if err != nil {
						fatalf("could not build logging configuration: %v\n\nTo see possible logging "+
							"instructions run:\n  burrow configure --describe-logging", err)
					}
					conf.Logging = &logging_config.LoggingConfig{
						RootSink: sinkConfig,
					}
				}

				if *jsonOutOpt {
					os.Stdout.WriteString(conf.JSONString())
				} else {
					os.Stdout.WriteString(conf.TOMLString())
				}
			}
		})

	burrow.Command("help",
		"Get more detailed or exhaustive options of selected commands or flags.",
		func(cmd *cli.Cmd) {

			cmd.Spec = "[--participant-accounts] [--full-accounts] [--toml]"

			cmd.Action = func() {
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

func burrowConfigProvider(configFile string) source.ConfigProvider {
	return source.FirstOf(
		// Will fail if file doesn't exist, but still skipped it configFile == ""
		source.TOMLFile(configFile, false),
		source.Environment(config.DefaultBurrowConfigJSONEnvironmentVariable),
		// Try working directory
		source.TOMLFile(config.DefaultBurrowConfigTOMLFileName, true),
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
