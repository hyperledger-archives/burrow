package commands

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/deployment"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/logging"
	logging_config "github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/config/presets"
	"github.com/jawher/mow.cli"
)

func Configure(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		genesisSpecOpt := cmd.StringOpt("s genesis-spec", "",
			"A GenesisSpec to use as a template for a GenesisDoc that will be created along with keys")

		jsonOutOpt := cmd.BoolOpt("j json", false, "Emit config in JSON rather than TOML "+
			"suitable for further processing")

		keysUrlOpt := cmd.StringOpt("k keys-url", "", fmt.Sprintf("Provide keys URL, default: %s",
			keys.DefaultKeysConfig().URL))

		configOpt := cmd.StringOpt("c base-config", "", "Use the a specified burrow config file as a base")

		genesisDocOpt := cmd.StringOpt("g genesis-doc", "", "GenesisDoc in JSON or TOML to embed in config")

		generateKeysOpt := cmd.StringOpt("x generate-keys", "",
			"File to output containing secret keys as JSON or according to a custom template (see --keys-template). "+
				"Note that using this options means the keys will not be generated in the default keys instance")

		keysTemplateOpt := cmd.StringOpt("z keys-template", deployment.DefaultDumpKeysFormat,
			fmt.Sprintf("Go text/template template (left delim: %s right delim: %s) to generate secret keys "+
				"file specified with --generate-keys.", deployment.LeftTemplateDelim, deployment.RightTemplateDelim))

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
			"[--separate-genesis-doc=<genesis JSON file>] [--chain-name] [--json] " +
			"[--logging=<logging program>] [--describe-logging] [--debug]"

		cmd.Action = func() {
			conf := config.DefaultBurrowConfig()

			if *configOpt != "" {
				// If explicitly given a config file use it as a base:
				err := source.FromFile(*configOpt, conf)
				if err != nil {
					output.Fatalf("could not read base config file (as TOML): %v", err)
				}
			}

			if *describeLoggingOpt {
				output.Logf("Usage:\n  burrow configure -l INSTRUCTION[,...]\n\nBuilds a logging " +
					"configuration by constructing a tree of logging sinks assembled from preset instructions " +
					"that generate the tree while traversing it.\n\nLogging Instructions:\n")
				for _, instruction := range presets.Instructons() {
					output.Logf("  %-15s\t%s\n", instruction.Name(), instruction.Description())
				}
				output.Logf("\nExample Usage:\n  burrow configure -l include-any,info,stderr\n")
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
					output.Fatalf("Could not read GenesisSpec: %v", err)
				}
				if *generateKeysOpt != "" {
					keyClient := mock.NewKeyClient()
					conf.GenesisDoc, err = genesisSpec.GenesisDoc(keyClient)
					if err != nil {
						output.Fatalf("Could not generate GenesisDoc from GenesisSpec using MockKeyClient: %v", err)
					}

					pkg := deployment.Package{Keys: keyClient.Keys()}
					secretKeysString, err := pkg.Dump(*keysTemplateOpt)
					if err != nil {
						output.Fatalf("Could not dump keys: %v", err)
					}
					err = ioutil.WriteFile(*generateKeysOpt, []byte(secretKeysString), 0700)
					if err != nil {
						output.Fatalf("Could not write secret keys: %v", err)
					}
				} else {
					conf.GenesisDoc, err = genesisSpec.GenesisDoc(keys.NewKeyClient(conf.Keys.URL, logging.NewNoopLogger()))
				}
				if err != nil {
					output.Fatalf("could not realise GenesisSpec: %v", err)
				}
			} else if *genesisDocOpt != "" {
				genesisDoc := new(genesis.GenesisDoc)
				err := source.FromFile(*genesisSpecOpt, genesisDoc)
				if err != nil {
					output.Fatalf("could not read GenesisSpec: %v", err)
				}
				conf.GenesisDoc = genesisDoc
			}

			// Logging
			if *loggingOpt != "" {
				ops := strings.Split(*loggingOpt, ",")
				sinkConfig, err := presets.BuildSinkConfig(ops...)
				if err != nil {
					output.Fatalf("could not build logging configuration: %v\n\nTo see possible logging "+
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
					output.Fatalf("Unable to set ChainName since no GenesisDoc/GenesisSpec provided.")
				}
				conf.GenesisDoc.ChainName = *chainNameOpt
			}

			if *separateGenesisDoc != "" {
				if conf.GenesisDoc == nil {
					output.Fatalf("Cannot write separate genesis doc since no GenesisDoc/GenesisSpec provided.")
				}
				genesisDocJSON, err := conf.GenesisDoc.JSONBytes()
				if err != nil {
					output.Fatalf("Could not form GenesisDoc JSON: %v", err)
				}
				err = ioutil.WriteFile(*separateGenesisDoc, genesisDocJSON, 0700)
				if err != nil {
					output.Fatalf("Could not write GenesisDoc JSON: %v", err)
				}
				conf.GenesisDoc = nil
			}
			if *jsonOutOpt {
				output.Printf(conf.JSONString())
			} else {
				output.Printf(conf.TOMLString())
			}
		}
	}
}
