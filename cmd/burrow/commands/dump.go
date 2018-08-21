package commands

import (
	"encoding/json"

	"github.com/hyperledger/burrow/forensics"
	"github.com/hyperledger/burrow/txs"
	cli "github.com/jawher/mow.cli"
	"github.com/tendermint/tendermint/libs/db"
)

func Dump(output Output) func(cmd *cli.Cmd) {
	return func(dump *cli.Cmd) {
		configOpt := dump.StringOpt("c config", "", "Use the a specified burrow config file")

		var explorer *forensics.BlockExplorer

		dump.Before = func() {
			conf, err := obtainBurrowConfig(*configOpt, "")
			if err != nil {
				output.Fatalf("Could not obtain config: %v", err)
			}
			tmConf := conf.Tendermint.TendermintConfig()

			explorer = forensics.NewBlockExplorer(db.DBBackendType(tmConf.DBBackend), tmConf.DBDir())
		}

		dump.Command("blocks", "dump blocks to stdout", func(cmd *cli.Cmd) {
			rangeArg := cmd.StringArg("RANGE", "", "Range as START_HEIGHT:END_HEIGHT where omitting "+
				"either endpoint implicitly describes the start/end and a negative index counts back from the last block")

			cmd.Spec = "[RANGE]"

			cmd.Action = func() {
				start, end, err := parseRange(*rangeArg)

				_, err = explorer.Blocks(start, end,
					func(block *forensics.Block) (stop bool) {
						bs, err := json.Marshal(block)
						if err != nil {
							output.Fatalf("Could not serialise block: %v", err)
						}
						output.Printf(string(bs))
						return false
					})
				if err != nil {
					output.Fatalf("Error iterating over blocks: %v", err)
				}
			}
		})

		dump.Command("txs", "dump transactions to stdout", func(cmd *cli.Cmd) {
			rangeArg := cmd.StringArg("RANGE", "", "Range as START_HEIGHT:END_HEIGHT where omitting "+
				"either endpoint implicitly describes the start/end and a negative index counts back from the last block")

			cmd.Spec = "[RANGE]"

			cmd.Action = func() {
				start, end, err := parseRange(*rangeArg)

				_, err = explorer.Blocks(start, end,
					func(block *forensics.Block) (stop bool) {
						stopped, err := block.Transactions(func(txEnv *txs.Envelope) (stop bool) {
							wrapper := struct {
								Height int64
								Tx     *txs.Envelope
							}{
								Height: block.Height,
								Tx:     txEnv,
							}
							bs, err := json.Marshal(wrapper)
							if err != nil {
								output.Fatalf("Could not deserialise transaction: %v", err)
							}
							output.Printf(string(bs))
							return false
						})
						if err != nil {
							output.Fatalf("Error iterating over transactions: %v", err)
						}
						// If we stopped transactions stop everything
						return stopped
					})
				if err != nil {
					output.Fatalf("Error iterating over blocks: %v", err)
				}
			}
		})
	}
}
