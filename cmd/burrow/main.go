package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/logging"
	logging_config "github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/config/presets"
	"github.com/hyperledger/burrow/project"
	"github.com/jawher/mow.cli"
)

func main() {
	burrow().Run(os.Args)
}

func burrow() *cli.Cli {
	app := cli.App("burrow", "The EVM smart contract machine with Tendermint consensus")

	versionOpt := app.BoolOpt("v version", false, "Print the Burrow version")
	app.Spec = "[--version]"

	app.Action = func() {
		if *versionOpt {
			fmt.Println(project.FullVersion())
			os.Exit(0)
		}
	}

	app.Command("serve", "",
		func(cmd *cli.Cmd) {
			genesisOpt := cmd.StringOpt("g genesis", "",
				"Use the specified genesis JSON file rather than a key in the main config, use - to read from STDIN")

			configOpt := cmd.StringOpt("c config", "", "Use the a specified burrow config file")

			validatorIndexOpt := cmd.Int(cli.IntOpt{
				Name:   "v validator-index",
				Desc:   "Validator index (in validators list - GenesisSpec or GenesisDoc) from which to set ValidatorAddress",
				Value:  -1,
				EnvVar: "BURROW_VALIDATOR_INDEX",
			})

			validatorAddressOpt := cmd.String(cli.StringOpt{
				Name:   "a validator-address",
				Desc:   "The address of the the signing key of this validator",
				EnvVar: "BURROW_VALIDATOR_ADDRESS",
			})

			validatorPassphraseOpt := cmd.String(cli.StringOpt{
				Name:   "p validator-passphrase",
				Desc:   "The passphrase of the signing key of this validator (currently unimplemented but planned for future version of our KeyClient interface)",
				EnvVar: "BURROW_VALIDATOR_PASSPHRASE",
			})

			cmd.Spec = "[--config=<config file>] " +
				"[--validator-index=<index of validator in GenesisDoc> | --validator-address=<address of validator signing key>] " +
				"[--genesis=<genesis json file>]"

			cmd.Action = func() {

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

				// Which validator am I?
				if *validatorAddressOpt != "" {
					address, err := acm.AddressFromHexString(*validatorAddressOpt)
					if err != nil {
						fatalf("could not read address for validator in '%s'", *validatorAddressOpt)
					}
					conf.ValidatorAddress = &address
				} else if *validatorIndexOpt > -1 {
					if conf.GenesisDoc == nil {
						fatalf("Unable to set ValidatorAddress from provided validator-index since no " +
							"GenesisDoc/GenesisSpec provided.")
					}
					if *validatorIndexOpt >= len(conf.GenesisDoc.Validators) {
						fatalf("validator-index of %v given but only %v validators specified in GenesisDoc",
							*validatorIndexOpt, len(conf.GenesisDoc.Validators))
					}
					conf.ValidatorAddress = &conf.GenesisDoc.Validators[*validatorIndexOpt].Address
					printf("Using validator index %v (address: %s)", *validatorIndexOpt, *conf.ValidatorAddress)
				}

				if *validatorPassphraseOpt != "" {
					conf.ValidatorPassphrase = validatorPassphraseOpt
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
		})

	app.Command("spec",
		"Build a GenesisSpec that acts as a template for a GenesisDoc and the configure command",
		func(cmd *cli.Cmd) {
			tomlOpt := cmd.BoolOpt("t toml", false, "Emit GenesisSpec as TOML rather than the "+
				"default JSON")

			baseOpt := cmd.StringsOpt("b base", nil, "Provide a base GenesisSpecs on top of which any "+
				"additional GenesisSpec presets specified by other flags will be merged. GenesisSpecs appearing "+
				"later take precedent over those appearing early if multiple --base flags are provided")

			fullOpt := cmd.IntOpt("f full-accounts", 1, "Number of preset Full type accounts")
			validatorOpt := cmd.IntOpt("v validator-accounts", 0, "Number of preset Validator type accounts")
			rootOpt := cmd.IntOpt("r root-accounts", 0, "Number of preset Root type accounts")
			developerOpt := cmd.IntOpt("d developer-accounts", 0, "Number of preset Developer type accounts")
			participantsOpt := cmd.IntOpt("p participant-accounts", 1, "Number of preset Participant type accounts")
			chainNameOpt := cmd.StringOpt("n chain-name", "", "Default chain name")

			cmd.Spec = "[--base][--full-accounts] [--validator-accounts] [--root-accounts] [--developer-accounts] " +
				"[--participant-accounts] [--chain-name] [--toml]"

			cmd.Action = func() {
				specs := make([]spec.GenesisSpec, 0, *participantsOpt+*fullOpt)
				for _, baseSpec := range *baseOpt {
					genesisSpec := new(spec.GenesisSpec)
					err := source.FromFile(baseSpec, genesisSpec)
					if err != nil {
						fatalf("could not read GenesisSpec: %v", err)
					}
					specs = append(specs, *genesisSpec)
				}
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
				if *chainNameOpt != "" {
					genesisSpec.ChainName = *chainNameOpt
				}
				if *tomlOpt {
					os.Stdout.WriteString(source.TOMLString(genesisSpec))
				} else {
					os.Stdout.WriteString(source.JSONString(genesisSpec))
				}
			}
		})

	app.Command("configure",
		"Create Burrow configuration by consuming a GenesisDoc or GenesisSpec, creating keys, and emitting the config",
		func(cmd *cli.Cmd) {
			genesisSpecOpt := cmd.StringOpt("s genesis-spec", "",
				"A GenesisSpec to use as a template for a GenesisDoc that will be created along with keys")

			jsonOutOpt := cmd.BoolOpt("j json-out", false, "Emit config in JSON rather than TOML "+
				"suitable for further processing or forming a separate genesis.json GenesisDoc")

			keysUrlOpt := cmd.StringOpt("k keys-url", "", fmt.Sprintf("Provide keys URL, default: %s",
				keys.DefaultKeysConfig().URL))

			configOpt := cmd.StringOpt("c base-config", "", "Use the a specified burrow config file as a base")

			genesisDocOpt := cmd.StringOpt("g genesis-doc", "", "GenesisDoc in JSON or TOML to embed in config")

			generateKeysOpt := cmd.StringOpt("x generate-keys", "",
				"File to output containing secret keys as JSON or according to a custom template (see --keys-template). "+
					"Note that using this options means the keys will not be generated in the default keys instance")

			keysTemplateOpt := cmd.StringOpt("z keys-template", mock.DefaultDumpKeysFormat,
				fmt.Sprintf("Go text/template template (left delim: %s right delim: %s) to generate secret keys "+
					"file specified with --generate-keys. Default:\n%s", mock.LeftTemplateDelim, mock.RightTemplateDelim,
					mock.DefaultDumpKeysFormat))

			separateGenesisDoc := cmd.StringOpt("w separate-genesis-doc", "", "Emit a separate genesis doc as JSON or TOML")

			loggingOpt := cmd.StringOpt("l logging", "",
				"Comma separated list of logging instructions which form a 'program' which is a depth-first "+
					"pre-order of instructions that will build the root logging sink. See 'burrow help' for more information.")

			describeLoggingOpt := cmd.BoolOpt("describe-logging", false,
				"Print an exhaustive list of logging instructions available with the --logging option")

			debugOpt := cmd.BoolOpt("d debug", false, "Include maximal debug options in config "+
				"including logging opcodes and dumping EVM tokens to disk these can be later pruned from the "+
				"generated config.")

			chainNameOpt := cmd.StringOpt("n chain-name", "", "Default chain name")

			cmd.Spec = "[--keys-url=<keys URL> | (--generate-keys=<secret keys files> [--keys-template=<text template for each key>])] " +
				"[--genesis-spec=<GenesisSpec file> | --genesis-doc=<GenesisDoc file>] " +
				"[--separate-genesis-doc=<genesis JSON file>] [--chain-name] [--json-out] " +
				"[--logging=<logging program>] [--describe-logging] [--debug]"

			cmd.Action = func() {
				conf := config.DefaultBurrowConfig()

				if *configOpt != "" {
					// If explicitly given a config file use it as a base:
					err := source.FromFile(*configOpt, conf)
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

				// Genesis Spec
				if *genesisSpecOpt != "" {
					genesisSpec := new(spec.GenesisSpec)
					err := source.FromFile(*genesisSpecOpt, genesisSpec)
					if err != nil {
						fatalf("Could not read GenesisSpec: %v", err)
					}
					if *generateKeysOpt != "" {
						keyClient := mock.NewMockKeyClient()
						conf.GenesisDoc, err = genesisSpec.GenesisDoc(keyClient)
						if err != nil {
							fatalf("Could not generate GenesisDoc from GenesisSpec using MockKeyClient: %v", err)
						}

						secretKeysString, err := keyClient.DumpKeys(*keysTemplateOpt)
						if err != nil {
							fatalf("Could not dump keys: %v", err)
						}
						err = ioutil.WriteFile(*generateKeysOpt, []byte(secretKeysString), 0700)
						if err != nil {
							fatalf("Could not write secret keys: %v", err)
						}
					} else {
						conf.GenesisDoc, err = genesisSpec.GenesisDoc(keys.NewKeyClient(conf.Keys.URL, logging.NewNoopLogger()))
					}
					if err != nil {
						fatalf("could not realise GenesisSpec: %v", err)
					}
				} else if *genesisDocOpt != "" {
					genesisDoc := new(genesis.GenesisDoc)
					err := source.FromFile(*genesisSpecOpt, genesisDoc)
					if err != nil {
						fatalf("could not read GenesisSpec: %v", err)
					}
					conf.GenesisDoc = genesisDoc
				}

				// Logging
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

				if *debugOpt {
					conf.Execution = &execution.ExecutionConfig{
						VMOptions: []execution.VMOption{execution.DumpTokens, execution.DebugOpcodes},
					}
				}

				if *chainNameOpt != "" {
					if conf.GenesisDoc == nil {
						fatalf("Unable to set ChainName since no GenesisDoc/GenesisSpec provided.")
					}
					conf.GenesisDoc.ChainName = *chainNameOpt
				}

				if *separateGenesisDoc != "" {
					if conf.GenesisDoc == nil {
						fatalf("Cannot write separate genesis doc since no GenesisDoc/GenesisSpec provided.")
					}
					genesisDocJSON, err := conf.GenesisDoc.JSONBytes()
					if err != nil {
						fatalf("Could not form GenesisDoc JSON: %v", err)
					}
					err = ioutil.WriteFile(*separateGenesisDoc, genesisDocJSON, 0700)
					if err != nil {
						fatalf("Could not write GenesisDoc JSON: %v", err)
					}
					conf.GenesisDoc = nil
				}
				if *jsonOutOpt {
					os.Stdout.WriteString(conf.JSONString())
				} else {
					os.Stdout.WriteString(conf.TOMLString())
				}
			}
		})

	app.Command("help",
		"Get more detailed or exhaustive options of selected commands or flags.",
		func(cmd *cli.Cmd) {

			cmd.Spec = "[--participant-accounts] [--full-accounts] [--toml]"

			cmd.Action = func() {
			}
		})

	return app
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
		source.File(configFile, false),
		source.Environment(config.DefaultBurrowConfigJSONEnvironmentVariable),
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
