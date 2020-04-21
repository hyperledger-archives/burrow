package commands

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/jobs"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs/payload"
	cli "github.com/jawher/mow.cli"
)

// Tx constructs or sends payloads to a burrow daemon
func Tx(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		configOpts := addConfigOptions(cmd)
		chainOpt := cmd.StringOpt("chain", "", "chain to be used in IP:PORT format")
		timeoutOpt := cmd.IntOpt("t timeout", 5, "Timeout in seconds")
		cmd.Spec += "[--chain=<ip>] [--timeout=<seconds>]"
		// we don't want config sourcing logs
		source.LogWriter = ioutil.Discard

		// formulate first to enable better visibility for the tx input
		cmd.Command("formulate", "formulate a tx", func(cmd *cli.Cmd) {
			conf, err := configOpts.obtainBurrowConfig()
			if err != nil {
				output.Fatalf("could not set up config: %v", err)
			}
			if err := conf.Verify(); err != nil {
				output.Fatalf("cannot continue with config: %v", err)
			}

			chainHost := jobs.FirstOf(*chainOpt, conf.RPC.GRPC.ListenAddress())
			client := def.NewClient(chainHost, conf.Keys.RemoteAddress, true, time.Duration(*timeoutOpt)*time.Second)
			logger := logging.NewNoopLogger()
			address := conf.ValidatorAddress.String()

			cmd.Command("send", "send value to another account", func(cmd *cli.Cmd) {
				sourceOpt := cmd.StringOpt("s source", "", "Address to send from, if not set config is used")
				targetOpt := cmd.StringOpt("t target", "", "Address to receive transfer, required")
				amountOpt := cmd.StringOpt("a amount", "", "Amount of value to send, required")
				cmd.Spec += "[--source=<address>] [--target=<address>] [--amount=<value>]"

				cmd.Action = func() {
					send := &def.Send{
						Source:      jobs.FirstOf(*sourceOpt, address),
						Destination: *targetOpt,
						Amount:      *amountOpt,
					}

					if err := send.Validate(); err != nil {
						output.Fatalf("could not validate SendTx: %v", err)
					}

					tx, err := jobs.FormulateSendJob(send, address, client, logger)
					if err != nil {
						output.Fatalf("could not formulate SendTx: %v", err)
					}

					output.Printf("%s", source.JSONString(payload.Any{
						SendTx: tx,
					}))
				}
			})

			cmd.Command("bond", "bond a new validator", func(cmd *cli.Cmd) {
				sourceOpt := cmd.StringOpt("s source", "", "Account with bonding perm, if not set config is used")
				amountOpt := cmd.StringOpt("a amount", "", "Amount of value to bond, required")
				cmd.Spec += "[--source=<address>] [--amount=<value>]"

				cmd.Action = func() {
					bond := &def.Bond{
						Source: jobs.FirstOf(*sourceOpt, address),
						Amount: *amountOpt,
					}

					if err := bond.Validate(); err != nil {
						output.Fatalf("could not validate BondTx: %v", err)
					}

					tx, err := jobs.FormulateBondJob(bond, address, client, logger)
					if err != nil {
						output.Fatalf("could not formulate BondTx: %v", err)
					}

					output.Printf("%s", source.JSONString(payload.Any{
						BondTx: tx,
					}))
				}
			})

			cmd.Command("unbond", "unbond an existing validator", func(cmd *cli.Cmd) {
				sourceOpt := cmd.StringOpt("s source", "", "Validator to unbond, if not set config is used")
				amountOpt := cmd.StringOpt("a amount", "", "Amount of value to unbond, required")
				cmd.Spec += "[--source=<address>] [--amount=<value>]"

				cmd.Action = func() {
					unbond := &def.Unbond{
						Source: jobs.FirstOf(*sourceOpt, address),
						Amount: *amountOpt,
					}

					if err := unbond.Validate(); err != nil {
						output.Fatalf("could not validate UnbondTx: %v", err)
					}

					tx, err := jobs.FormulateUnbondJob(unbond, address, client, logger)
					if err != nil {
						output.Fatalf("could not formulate UnbondTx: %v", err)
					}

					output.Printf("%s", source.JSONString(payload.Any{
						UnbondTx: tx,
					}))
				}
			})

			cmd.Command("identify", "associate a validator with a node address", func(cmd *cli.Cmd) {
				sourceOpt := cmd.StringOpt("source", "", "Address to send from, if not set config is used")
				nodeKeyOpt := cmd.StringOpt("node-key", "", "File containing the nodeKey to use, default config")
				networkOpt := cmd.StringOpt("network", "", "Publically reachable host IP")
				monikerOpt := cmd.StringOpt("moniker", "", "Human readable node ID")
				cmd.Spec += "[--source=<address>] [--node-key=<file>] [--network=<address>] [--moniker=<name>]"

				cmd.Action = func() {

					tmConf, err := conf.TendermintConfig()
					if err != nil {
						output.Fatalf("could not construct tendermint config: %v", err)
					}

					id := &def.Identify{
						Source:     jobs.FirstOf(*sourceOpt, address),
						NodeKey:    jobs.FirstOf(*nodeKeyOpt, tmConf.NodeKeyFile()),
						Moniker:    *monikerOpt,
						NetAddress: jobs.FirstOf(*networkOpt, conf.Tendermint.ListenHost),
					}

					if err := id.Validate(); err != nil {
						output.Fatalf("could not validate IdentifyTx: %v", err)
					}

					tx, err := jobs.FormulateIdentifyJob(id, address, client, logger)
					if err != nil {
						output.Fatalf("could not formulate IdentifyTx: %v", err)
					}

					output.Printf("%s", source.JSONString(payload.Any{
						IdentifyTx: tx,
					}))
				}
			})
		})

		cmd.Command("commit", "read and send a tx to mempool", func(cmd *cli.Cmd) {
			conf, err := configOpts.obtainBurrowConfig()
			if err != nil {
				output.Fatalf("could not set up config: %v", err)
			}
			fileOpt := cmd.StringOpt("f file", "", "Read the tx spec from a file")
			cmd.Spec += "[--file=<location>]"

			cmd.Action = func() {
				if err := conf.Verify(); err != nil {
					output.Fatalf("can't continue with config: %v", err)
				}

				chainHost := jobs.FirstOf(*chainOpt, conf.RPC.GRPC.ListenAddress())
				client := def.NewClient(chainHost, conf.Keys.RemoteAddress, true, time.Duration(*timeoutOpt)*time.Second)

				var rawTx payload.Any
				var hash string
				var err error

				data, err := readInput(*fileOpt)
				if err != nil {
					output.Fatalf("no input: %v", err)
				}

				if err = json.Unmarshal(data, &rawTx); err != nil {
					output.Fatalf("could not unmarshal Tx: %v", err)
				}

				switch tx := rawTx.GetValue().(type) {
				case *payload.SendTx:
					hash, err = makeTx(client, tx)
				case *payload.BondTx:
					hash, err = makeTx(client, tx)
				case *payload.UnbondTx:
					hash, err = makeTx(client, tx)
				case *payload.IdentifyTx:
					hash, err = makeTx(client, tx)
				default:
					output.Fatalf("payload type not recognized")
				}

				if err != nil {
					output.Fatalf("failed to commit tx to mempool: %v", err)
				}

				output.Printf("%s", hash)
			}
		})
	}
}

func makeTx(client *def.Client, tx payload.Payload) (string, error) {
	logger := logging.NewNoopLogger()
	txe, err := client.SignAndBroadcast(tx, logger)
	if err != nil {
		return "", err
	}

	jobs.LogTxExecution(txe, logger)
	if err != nil {
		return "", err
	}
	return txe.Receipt.TxHash.String(), nil
}

func readInput(file string) ([]byte, error) {
	if file != "" {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, errors.New("expected input from STDIN, use --file otherwise")
	}

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, errors.New("could not read data from STDIN")
	}

	return data, nil
}
