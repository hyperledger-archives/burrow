package commands

import (
	"github.com/eris-ltd/eris-db/genesis"

	"github.com/spf13/cobra"
)

var GenesisGenCmd = &cobra.Command{
	Use:   "genesis",
	Short: "eris-client genesis creates a genesis.json with known inputs",
	Long:  "eris-client genesis creates a genesis.json with known inputs",

	Run: func(cmd *cobra.Command, args []string) {

		genesis.GenerateKnown("thisIsChainID", "", "114234767676")

	},
}

func buildGenesisGenCommand() {
	//addTransactionPersistentFlags()
}

var (
	DirFlag string
	//AddrsFlag  string
	CsvPathFlag       string
	PubkeyFlag        string
	RootFlag          string
	NoValAccountsFlag bool
)

//var knownCmd = &cobra.Command{
//	Use:   "known",
//	Short: "mintgen known <chain_id> [flags] ",
//	Long:  "Create a genesis.json with --pub <pub_1> <pub_2> <pub_N> or with --csv <path_to_file>, or pass a priv_validator.json on stdin. Two csv file names can be passed (comma separated) to distinguish validators and accounts.",
//	//Run:   cliKnown,
//}
//
//
//knownCmd.Flags().StringVarP(&PubkeyFlag, "pub", "", "", "pubkeys to include when generating genesis.json. flag is req'd")
//knownCmd.Flags().StringVarP(&CsvPathFlag, "csv", "", "", "path to .csv with the following params: (pubkey, starting balance, name, permissions, setbit")
//
