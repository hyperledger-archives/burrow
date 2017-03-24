// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"

	"github.com/monax/eris-db/common/sanity"
	"github.com/monax/eris-db/genesis"

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
