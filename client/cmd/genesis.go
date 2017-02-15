package commands

import (
	"fmt"

	"github.com/eris-ltd/eris-db/common/sanity"
	"github.com/eris-ltd/eris-db/genesis"

	"github.com/spf13/cobra"
)

// TODO refactor these vars into a struct?
var (
	AccountsPathFlag   string
	ValidatorsPathFlag string
)

var GenesisGenCmd = &cobra.Command{
	Use:   "make-genesis",
	Short: "eris-client make-genesis creates a genesis.json with known inputs",
	Long:  "eris-client make-genesis creates a genesis.json with known inputs",

	Run: func(cmd *cobra.Command, args []string) {
		// TODO refactor to not panic
		genesisFile, err := genesis.GenerateKnown(args[0], AccountsPathFlag, ValidatorsPathFlag)
		if err != nil {
			sanity.PanicSanity(err)
		}
		fmt.Println(genesisFile) // may want to save somewhere instead
	},
}

func buildGenesisGenCommand() {
	addGenesisPersistentFlags()
}

func addGenesisPersistentFlags() {
	GenesisGenCmd.Flags().StringVarP(&AccountsPathFlag, "accounts", "", "", "path to accounts.csv with the following params: (pubkey, starting balance, name, permissions, setbit")
	GenesisGenCmd.Flags().StringVarP(&ValidatorsPathFlag, "validators", "", "", "path to validators.csv with the following params: (pubkey, starting balance, name, permissions, setbit")
}
