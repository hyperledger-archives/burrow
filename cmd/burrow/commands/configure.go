package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/deployment"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	logging_config "github.com/hyperledger/burrow/logging/config"
	"github.com/hyperledger/burrow/logging/config/presets"
	"github.com/jawher/mow.cli"
	"github.com/tendermint/go-amino"
	tm_crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tendermint/p2p"
)

func processTemplate(config *genesis.GenesisDoc, templateIn, templateOut string) error {
	data, err := ioutil.ReadFile(templateIn)
	if err != nil {
		return err
	}
	pkg := deployment.Config{Config: *config}
	output, err := pkg.Dump(string(data))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(templateOut, []byte(output), 0644)
}

func Configure(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		genesisSpecOpt := cmd.StringOpt("s genesis-spec", "",
			"A GenesisSpec to use as a template for a GenesisDoc that will be created along with keys")

		jsonOutOpt := cmd.BoolOpt("j json", false, "Emit config in JSON rather than TOML "+
			"suitable for further processing")

		keysUrlOpt := cmd.StringOpt("k keys-url", "", fmt.Sprintf("Provide keys GRPC address, default: %s",
			keys.DefaultKeysConfig().RemoteAddress))

		configOpt := cmd.StringOpt("c base-config", "", "Use the a specified burrow config file as a base")

		genesisDocOpt := cmd.StringOpt("g genesis-doc", "", "GenesisDoc in JSON or TOML to embed in config")

		generateNodeKeys := cmd.BoolOpt("generate-node-keys", false, "Generate node keys for validators")

		generateKeysOpt := cmd.StringOpt("x generate-keys", "",
			"File to output keys in a kubernetes secret definition (see --kubernetes-template)")

		keysTemplateOpt := cmd.StringOpt("z keys-template", deployment.KubernetesKeyDumpFormat,
			fmt.Sprintf("Go text/template template (left delim: %s right delim: %s) to generate secret keys "+
				"file specified with --generate-keys.", deployment.LeftTemplateDelim, deployment.RightTemplateDelim))

		configTemplateIn := cmd.StringsOpt("t config-template-in", nil,
			fmt.Sprintf("Go text/template template input filename (left delim: %s right delim: %s) to output generate config "+
				"file specified with --config-template-out", deployment.LeftTemplateDelim, deployment.RightTemplateDelim))

		configTemplateOut := cmd.StringsOpt("t config-template-out", nil,
			"Go text/template template output file. Template filename specified with --config-template-in "+
				"file specified with --config-template-out")

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

		cmd.Spec = "[--keys-url=<keys URL> | (--generate-keys=<secret keys files> [--keys-template=<text template for key>])] " +
			"[--config-template-in=<text template> --config-template-out=<output file>] " +
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
				conf.Keys.RemoteAddress = *keysUrlOpt
			}

			if len(*configTemplateIn) != len(*configTemplateOut) {
				output.Fatalf("--config-template-in and --config-template-out must be specified the same number of times")
			}

			// Genesis Spec
			if *genesisSpecOpt != "" {
				genesisSpec := new(spec.GenesisSpec)
				err := source.FromFile(*genesisSpecOpt, genesisSpec)
				if err != nil {
					output.Fatalf("Could not read GenesisSpec: %v", err)
				}
				keyStore := keys.NewKeyStore(conf.Keys.KeysDirectory, conf.Keys.AllowBadFilePermissions, logging.NewNoopLogger())

				if *generateKeysOpt != "" {
					keyClient := keys.NewLocalKeyClient(keyStore, logging.NewNoopLogger())
					conf.GenesisDoc, err = genesisSpec.GenesisDoc(keyClient, *generateNodeKeys)
					if err != nil {
						output.Fatalf("Could not generate GenesisDoc from GenesisSpec using MockKeyClient: %v", err)
					}

					allNames, err := keyStore.GetAllNames()
					if err != nil {
						output.Fatalf("could get all keys: %v", err)
					}

					cdc := amino.NewCodec()
					tm_crypto.RegisterAmino(cdc)

					pkg := deployment.KeysSecret{ChainName: conf.GenesisDoc.ChainName}

					for k := range allNames {
						addr, err := crypto.AddressFromHexString(allNames[k])
						if err != nil {
							output.Fatalf("Address %s not valid: %v", k, err)
						}
						key, err := keyStore.GetKey("", addr[:])
						if err != nil {
							output.Fatalf("Failed to get key: %s: %v", k, err)
						}

						// Is this is a validator node key?
						nodeKey := false
						for _, a := range conf.GenesisDoc.Validators {
							if a.NodeAddress != nil && addr == *a.NodeAddress {
								nodeKey = true
								break
							}
						}

						if nodeKey {
							privKey := tm_crypto.GenPrivKeyEd25519()
							copy(privKey[:], key.PrivateKey.PrivateKey)
							nodeKey := &p2p.NodeKey{
								PrivKey: privKey,
							}

							json, err := cdc.MarshalJSON(nodeKey)
							if err != nil {
								output.Fatalf("go-amino failed to json marshall private key: %v", err)
							}
							pkg.NodeKeys = append(pkg.NodeKeys, deployment.Key{Name: k, Address: addr, KeyJSON: json})
						} else {
							json, err := json.Marshal(key)
							if err != nil {
								output.Fatalf("Failed to json marshal key: %s: %v", k, err)
							}
							pkg.Keys = append(pkg.Keys, deployment.Key{Name: k, Address: addr, KeyJSON: json})
						}
					}

					secretKeysString, err := pkg.Dump(*keysTemplateOpt)
					if err != nil {
						output.Fatalf("Could not dump keys: %v", err)
					}
					err = ioutil.WriteFile(*generateKeysOpt, []byte(secretKeysString), 0700)
					if err != nil {
						output.Fatalf("Could not write secret keys: %v", err)
					}
				} else {
					var keyClient keys.KeyClient
					if conf.Keys.RemoteAddress != "" {
						keyClient, err = keys.NewRemoteKeyClient(conf.Keys.RemoteAddress, logging.NewNoopLogger())
						if err != nil {
							output.Fatalf("Could not create remote key client: %v", err)
						}
					} else {
						keyClient = keys.NewLocalKeyClient(keyStore, logging.NewNoopLogger())
					}
					conf.GenesisDoc, err = genesisSpec.GenesisDoc(keyClient, *generateNodeKeys)
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

			for ind := range *configTemplateIn {
				err := processTemplate(conf.GenesisDoc, (*configTemplateIn)[ind], (*configTemplateOut)[ind])
				if err != nil {
					output.Fatalf("coult not template from %s to %s: %v", (*configTemplateIn)[ind], (*configTemplateOut)[ind], err)
				}
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
