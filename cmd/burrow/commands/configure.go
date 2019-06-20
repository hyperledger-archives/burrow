package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hyperledger/burrow/consensus/tendermint"

	"github.com/hyperledger/burrow/dump"

	"github.com/hyperledger/burrow/config/deployment"
	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/logconfig"
	"github.com/hyperledger/burrow/logging/logconfig/presets"
	"github.com/hyperledger/burrow/rpc"
	cli "github.com/jawher/mow.cli"
	amino "github.com/tendermint/go-amino"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/libs/db"
)

// Configure generates burrow configuration(s)
func Configure(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {

		genesisSpecOpt := cmd.StringOpt("s genesis-spec", "",
			"A GenesisSpec to use as a template for a GenesisDoc that will be created along with keys")

		jsonOutOpt := cmd.BoolOpt("j json", false, "Emit config in JSON rather than TOML "+
			"suitable for further processing")

		keysURLOpt := cmd.StringOpt("k keys-url", "", fmt.Sprintf("Provide keys GRPC address, default: %s",
			keys.DefaultKeysConfig().RemoteAddress))

		keysDir := cmd.StringOpt("keys-dir", "", "Directory where keys are stored")

		configTemplateIn := cmd.StringsOpt("config-template-in", nil,
			fmt.Sprintf("Go text/template input filename to generate config file specified with --config-out"))

		configTemplateOut := cmd.StringsOpt("config-out", nil,
			"Go text/template output file. Template filename specified with --config-template-in")

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

		emptyBlocksOpt := cmd.StringOpt("e empty-blocks", "",
			"Whether to create empty blocks, one of: 'never' (always wait for transactions before proposing a "+
				"block, 'always' (at end of each consensus round), or a duration like '1s', '5m', or '6h'")

		restoreDumpOpt := cmd.StringOpt("restore-dump", "", "Including AppHash for restored file")

		pool := cmd.BoolOpt("pool", false, "Write config files for all the validators called burrowNNN.toml")

		cmd.Spec = "[--keys-url=<keys URL> | --keys-dir=<keys directory>] " +
			"[ --config-template-in=<text template> --config-out=<output file>]... " +
			"[--genesis-spec=<GenesisSpec file>] [--separate-genesis-doc=<genesis JSON file>] " +
			"[--chain-name=<chain name>] [--restore-dump=<dump file>] [--json] [--debug] [--pool] " +
			"[--logging=<logging program>] [--describe-logging] [--empty-blocks=<'always','never',duration>]"

		// no sourcing logs
		source.LogWriter = ioutil.Discard
		// TODO: this has default, only set if explicit?
		configOpts := addConfigOptions(cmd)

		cmd.Action = func() {
			conf, err := configOpts.obtainBurrowConfig()
			if err != nil {
				output.Fatalf("could not obtain config: %v", err)
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

			if *keysURLOpt != "" {
				conf.Keys.RemoteAddress = *keysURLOpt
			}

			if len(*configTemplateIn) != len(*configTemplateOut) {
				output.Fatalf("--config-template-in and --config-out must be specified the same number of times")
			}

			pkg := deployment.Config{}

			// Genesis Spec
			if *genesisSpecOpt != "" {
				genesisSpec := new(spec.GenesisSpec)
				err := source.FromFile(*genesisSpecOpt, genesisSpec)
				if err != nil {
					output.Fatalf("Could not read GenesisSpec: %v", err)
				}
				if conf.Keys.RemoteAddress == "" {
					dir := conf.Keys.KeysDirectory
					if *keysDir != "" {
						dir = *keysDir
					}
					keyStore := keys.NewKeyStore(dir, conf.Keys.AllowBadFilePermissions)

					keyClient := keys.NewLocalKeyClient(keyStore, logging.NewNoopLogger())
					conf.GenesisDoc, err = genesisSpec.GenesisDoc(keyClient)
					if err != nil {
						output.Fatalf("could not generate GenesisDoc from GenesisSpec using MockKeyClient: %v", err)
					}

					allNames, err := keyStore.GetAllNames()
					if err != nil {
						output.Fatalf("could get all keys: %v", err)
					}

					pkg = deployment.Config{Keys: make(map[crypto.Address]deployment.Key)}

					for k := range allNames {
						addr, err := crypto.AddressFromHexString(allNames[k])
						if err != nil {
							output.Fatalf("address %s not valid: %v", k, err)
						}
						key, err := keyStore.GetKey("", addr[:])
						if err != nil {
							output.Fatalf("failed to get key: %s: %v", k, err)
						}
						json, err := json.Marshal(key)
						if err != nil {
							output.Fatalf("failed to json marshal key: %s: %v", k, err)
						}
						pkg.Keys[addr] = deployment.Key{Name: k, Address: addr, KeyJSON: json}
					}
				} else {
					keyClient, err := keys.NewRemoteKeyClient(conf.Keys.RemoteAddress, logging.NewNoopLogger())
					if err != nil {
						output.Fatalf("could not create remote key client: %v", err)
					}
					conf.GenesisDoc, err = genesisSpec.GenesisDoc(keyClient)
					if err != nil {
						output.Fatalf("could not realise GenesisSpec: %v", err)
					}
				}

			}

			if *chainNameOpt != "" {
				if conf.GenesisDoc == nil {
					output.Fatalf("unable to set ChainName since no GenesisDoc/GenesisSpec provided.")
				}
				conf.GenesisDoc.ChainName = *chainNameOpt
			}

			if *restoreDumpOpt != "" {
				if conf.GenesisDoc == nil {
					output.Fatalf("no GenesisDoc provided, cannot restore dump")
				}

				if len(conf.GenesisDoc.Validators) == 0 {
					output.Fatalf("on restore, validators must be provided in GenesisDoc or GenesisSpec")
				}

				reader, err := dump.NewFileReader(*restoreDumpOpt)
				if err != nil {
					output.Fatalf("failed to read restore dump: %v", err)
				}

				st, err := state.MakeGenesisState(db.NewMemDB(), conf.GenesisDoc)
				if err != nil {
					output.Fatalf("could not generate state from genesis: %v", err)
				}

				err = dump.Load(reader, st)
				if err != nil {
					output.Fatalf("could not restore dump %s: %v", *restoreDumpOpt, err)
				}

				conf.GenesisDoc.AppHash = st.Hash()
			}

			if conf.GenesisDoc != nil {
				pkg.GenesisDoc = conf.GenesisDoc

				for _, v := range conf.GenesisDoc.Validators {
					nodeKey := tendermint.NewNodeKey()
					nodeAddress, _ := crypto.AddressFromHexString(string(nodeKey.ID()))

					cdc := amino.NewCodec()
					cryptoAmino.RegisterAmino(cdc)
					json, err := cdc.MarshalJSON(nodeKey)
					if err != nil {
						output.Fatalf("go-amino failed to json marshall private key: %v", err)
					}
					pkg.Keys[nodeAddress] = deployment.Key{Name: v.Name, Address: nodeAddress, KeyJSON: json}

					pkg.Validators = append(pkg.Validators, deployment.Validator{
						Name:        v.Name,
						Address:     v.Address,
						NodeAddress: nodeAddress,
					})
				}

				for ind := range *configTemplateIn {
					err := processTemplate(&pkg, (*configTemplateIn)[ind], (*configTemplateOut)[ind])
					if err != nil {
						output.Fatalf("could not template from %s to %s: %v", (*configTemplateIn)[ind], (*configTemplateOut)[ind], err)
					}
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
				conf.Logging = &logconfig.LoggingConfig{
					RootSink: sinkConfig,
				}
			}

			if *debugOpt {
				conf.Execution = &execution.ExecutionConfig{
					VMOptions: []execution.VMOption{execution.DumpTokens, execution.DebugOpcodes},
				}
			}

			if *separateGenesisDoc != "" {
				if conf.GenesisDoc == nil {
					output.Fatalf("cannot write separate genesis doc since no GenesisDoc/GenesisSpec was provided")
				}
				genesisDocJSON, err := conf.GenesisDoc.JSONBytes()
				if err != nil {
					output.Fatalf("could not form GenesisDoc JSON: %v", err)
				}
				err = ioutil.WriteFile(*separateGenesisDoc, genesisDocJSON, 0644)
				if err != nil {
					output.Fatalf("could not write GenesisDoc JSON: %v", err)
				}
				conf.GenesisDoc = nil
			}

			if *emptyBlocksOpt != "" {
				conf.Tendermint.CreateEmptyBlocks = *emptyBlocksOpt
			}

			if *pool {
				peers := make([]string, 0)
				for i := range conf.GenesisDoc.Validators {
					tmConf, err := conf.Tendermint.Config(fmt.Sprintf(".burrow%03d", i), conf.Execution.TimeoutFactor)
					if err != nil {
						output.Fatalf("could not obtain config for %03d: %v", i, err)
					}
					nodeKey, err := tendermint.EnsureNodeKey(tmConf.NodeKeyFile())
					if err != nil {
						output.Fatalf("failed to create node key for %03d: %v", i, err)
					}
					peers = append(peers, fmt.Sprintf("tcp://%s@127.0.0.1:%d", nodeKey.ID(), 26656+i))
				}
				for i, acc := range conf.GenesisDoc.Accounts {
					// set stuff
					conf.Address = &acc.Address
					conf.Tendermint.PersistentPeers = strings.Join(peers, ",")

					conf.BurrowDir = fmt.Sprintf(".burrow%03d", i)
					conf.Tendermint.ListenHost = rpc.LocalHost
					conf.Tendermint.ListenPort = fmt.Sprint(26656 + i)
					conf.RPC.Info.ListenHost = rpc.LocalHost
					conf.RPC.Info.ListenPort = fmt.Sprint(26758 + i)
					conf.RPC.GRPC.ListenHost = rpc.LocalHost
					conf.RPC.GRPC.ListenPort = fmt.Sprint(10997 + i)
					conf.RPC.Metrics.ListenHost = rpc.LocalHost
					conf.RPC.Metrics.ListenPort = fmt.Sprint(9102 + i)
					conf.Logging.RootSink.Output.OutputType = "file"
					conf.Logging.RootSink.Output.FileConfig = &logconfig.FileConfig{Path: fmt.Sprintf("burrow%03d.log", i)}

					if *jsonOutOpt {
						ioutil.WriteFile(fmt.Sprintf("burrow%03d.json", i), []byte(conf.JSONString()), 0644)
					} else {
						ioutil.WriteFile(fmt.Sprintf("burrow%03d.toml", i), []byte(conf.TOMLString()), 0644)
					}
				}
			} else if *jsonOutOpt {
				output.Printf(conf.JSONString())
			} else {
				output.Printf(conf.TOMLString())
			}

		}
	}
}

func processTemplate(pkg *deployment.Config, templateIn, templateOut string) error {
	data, err := ioutil.ReadFile(templateIn)
	if err != nil {
		return err
	}
	fmt.Println(templateIn)
	output, err := pkg.Dump(templateIn, string(data))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(templateOut, []byte(output), 0644)
}
