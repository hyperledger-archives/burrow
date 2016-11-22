package commands

import (
	"github.com/eris-ltd/eris-db/genesis"

	"github.com/spf13/cobra"
)

// TODO refactor these vars into a struct?
var (
	//DirFlag string
	//AddrsFlag  string
	AccountsPathFlag   string
	ValidatorsPathFlag string
	//CsvPathFlag        string
	//PubkeyFlag         string
	//RootFlag           string
	//NoValAccountsFlag  bool
)

var GenesisGenCmd = &cobra.Command{
	Use:   "genesis",
	Short: "eris-client genesis creates a genesis.json with known inputs",
	Long:  "eris-client genesis creates a genesis.json with known inputs",

	Run: func(cmd *cobra.Command, args []string) {

		genesis.GenerateKnown(args[0], AccountsPathFlag, ValidatorsPathFlag)

	},
}

func buildGenesisGenCommand() {
	addGenesisPersistentFlags()
}

func addGenesisPersistentFlags() {
	GenesisGenCmd.Flags().StringVarP(&AccountsPathFlag, "accounts", "", "", "path to accounts.csv with the following params: (pubkey, starting balance, name, permissions, setbit")
	GenesisGenCmd.Flags().StringVarP(&ValidatorsPathFlag, "validators", "", "", "path to validators.csv with the following params: (pubkey, starting balance, name, permissions, setbit")
}
