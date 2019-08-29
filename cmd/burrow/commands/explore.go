package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hyperledger/burrow/config"
	"github.com/hyperledger/burrow/forensics"

	"github.com/hyperledger/burrow/bcm"

	"github.com/hyperledger/burrow/txs"
	cli "github.com/jawher/mow.cli"
	dbm "github.com/tendermint/tm-db"
)

// Explore chain state(s)
func Explore(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		configOpts := addConfigOptions(cmd)
		var conf *config.BurrowConfig
		var explorer *bcm.BlockStore
		var err error

		cmd.Before = func() {
			conf, err = configOpts.obtainBurrowConfig()
			if err != nil {
				output.Fatalf("could not obtain config: %v", err)
			}
			tmConf, err := conf.TendermintConfig()
			if err != nil {
				output.Fatalf("could not build Tendermint config:", err)
			}

			if conf.GenesisDoc == nil {
				output.Fatalf("genesis doc is required")
			}

			explorer = bcm.NewBlockExplorer(dbm.DBBackendType(tmConf.DBBackend), tmConf.DBDir())
		}

		cmd.Command("dump", "pretty print the state tree at the given height", func(cmd *cli.Cmd) {
			heightOpt := cmd.IntOpt("height", 0, "The height to read, defaults to latest")
			stateDir := cmd.StringArg("STATE", "", "Directory containing burrow state")
			cmd.Spec = "[--height] [STATE]"

			cmd.Before = func() {
				if err := isDir(*stateDir); err != nil {
					output.Fatalf("could not obtain state: %v", err)
				}
			}

			cmd.Action = func() {
				replay := forensics.NewReplayFromDir(conf.GenesisDoc, *stateDir)
				height := uint64(*heightOpt)
				if height == 0 {
					height, err = replay.LatestHeight()
					if err != nil {
						output.Fatalf("could not read latest height: %v", err)
					}
				}
				err := replay.LoadAt(height)
				if err != nil {
					output.Fatalf("could not load state: %v", err)
				}

				fmt.Println(replay.State.Dump())
			}
		})

		cmd.Command("compare", "diff the state of two .burrow directories", func(cmd *cli.Cmd) {
			goodDir := cmd.StringArg("GOOD", "", "Directory containing expected state")
			badDir := cmd.StringArg("BAD", "", "Directory containing invalid state")
			heightOpt := cmd.IntOpt("height", 0, "The height to read, defaults to latest")
			cmd.Spec = "[--height] [GOOD] [BAD]"

			cmd.Before = func() {
				if err := isDir(*goodDir); err != nil {
					output.Fatalf("could not obtain state: %v", err)
				}
				if err := isDir(*badDir); err != nil {
					output.Fatalf("could not obtain state: %v", err)
				}
			}

			cmd.Action = func() {
				replay1 := forensics.NewReplayFromDir(conf.GenesisDoc, *goodDir)
				replay2 := forensics.NewReplayFromDir(conf.GenesisDoc, *badDir)

				h1, err := replay1.LatestHeight()
				if err != nil {
					output.Fatalf("could not get height for first replay: %v", err)
				}
				h2, err := replay2.LatestHeight()
				if err != nil {
					output.Fatalf("could not get height for second replay: %v", err)
				}

				height := h1
				if *heightOpt != 0 {
					height = uint64(*heightOpt)
				} else if h2 < h1 {
					height = h2
					output.Printf("States do not agree on last height, using min: %d", h2)
				} else {
					output.Printf("Using default last height: %d", h1)
				}

				recap1, err := replay1.Blocks(1, height)
				if err != nil {
					output.Fatalf("could not replay first state: %v", err)
				}

				recap2, err := replay2.Blocks(1, height)
				if err != nil {
					output.Fatalf("could not replay second state: %v", err)
				}

				if height, err := forensics.CompareCaptures(recap1, recap2); err != nil {
					output.Printf("difference in capture: %v", err)
					// TODO: compare at every height?
					if err := forensics.CompareStateAtHeight(replay1.State, replay2.State, height); err != nil {
						output.Fatalf("difference in state: %v", err)
					}
				}

				output.Printf("States match!")
			}
		})

		cmd.Command("blocks", "dump blocks to stdout", func(cmd *cli.Cmd) {
			rangeArg := cmd.StringArg("RANGE", "", "Range as START_HEIGHT:END_HEIGHT where omitting "+
				"either endpoint implicitly describes the start/end and a negative index counts back from the last block")

			cmd.Spec = "[RANGE]"

			cmd.Action = func() {
				start, end, err := parseRange(*rangeArg)
				if err != nil {
					output.Fatalf("could not parse range '%s': %v", *rangeArg, err)
				}

				err = explorer.Blocks(start, end,
					func(block *bcm.Block) error {
						bs, err := json.Marshal(block)
						if err != nil {
							output.Fatalf("Could not serialise block: %v", err)
						}
						output.Printf(string(bs))
						return nil
					})
				if err != nil {
					output.Fatalf("Error iterating over blocks: %v", err)
				}
			}
		})

		cmd.Command("txs", "dump transactions to stdout", func(cmd *cli.Cmd) {
			rangeArg := cmd.StringArg("RANGE", "", "Range as START_HEIGHT:END_HEIGHT where omitting "+
				"either endpoint implicitly describes the start/end and a negative index counts back from the last block")

			cmd.Spec = "[RANGE]"

			cmd.Action = func() {
				start, end, err := parseRange(*rangeArg)
				if err != nil {
					output.Fatalf("could not parse range '%s': %v", *rangeArg, err)
				}

				err = explorer.Blocks(start, end,
					func(block *bcm.Block) error {
						err := block.Transactions(func(txEnv *txs.Envelope) error {
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
							return nil
						})
						if err != nil {
							output.Fatalf("Error iterating over transactions: %v", err)
						}
						// If we stopped transactions stop everything
						return nil
					})
				if err != nil {
					output.Fatalf("Error iterating over blocks: %v", err)
				}
			}
		})
	}
}

func isDir(path string) error {
	file, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("could not read state directory: %v", err)
	} else if !file.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	return nil
}
