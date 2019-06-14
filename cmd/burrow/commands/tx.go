package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hyperledger/burrow/config/source"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/jobs"
	"github.com/hyperledger/burrow/deploy/util"
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

			chainHost := jobs.FirstOf(*chainOpt, fmt.Sprintf("%s:%s", conf.RPC.GRPC.ListenHost, conf.RPC.GRPC.ListenPort))
			client := def.NewClient(chainHost, conf.Keys.RemoteAddress, true, time.Duration(*timeoutOpt)*time.Second)
			logger := logging.NewNoopLogger()
			address := conf.Address.String()

			cmd.Command("send", "send value to another account", func(cmd *cli.Cmd) {
				sourceOpt := cmd.StringOpt("source", "", "Address to send from, if not set config is used")
				targetOpt := cmd.StringOpt("target", "", "Address to receive transfer, required")
				amountOpt := cmd.StringOpt("amount", "", "Amount of value to send, required")
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
		})

		cmd.Command("commit", "read and send a tx to mempool", func(cmd *cli.Cmd) {
			configOpts := addConfigOptions(cmd)
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

				chainHost := jobs.FirstOf(*chainOpt, fmt.Sprintf("%s:%s", conf.RPC.GRPC.ListenHost, conf.RPC.GRPC.ListenPort))
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

	util.ReadTxSignAndBroadcast(txe, err, logger)
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
