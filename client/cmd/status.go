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
	"github.com/spf13/cobra"

	"github.com/hyperledger/burrow/client/methods"
	"github.com/hyperledger/burrow/util"
)

func buildStatusCommand() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "burrow-client status returns the current status from a chain.",
		Long: `burrow-client status returns the current status from a chain.
`,
		Run: func(cmd *cobra.Command, args []string) {
			err := methods.Status(clientDo)
			if err != nil {
				util.Fatalf("Could not get status: %s", err)
			}
		},
	}
	statusCmd.PersistentFlags().StringVarP(&clientDo.NodeAddrFlag, "node-addr", "", defaultNodeRpcAddress(), "set the burrow node rpc server address (default respects $BURROW_CLIENT_NODE_ADDRESS)")
	// TransactionCmd.PersistentFlags().StringVarP(&clientDo.PubkeyFlag, "pubkey", "", defaultPublicKey(), "specify the public key to sign with (defaults to $BURROW_CLIENT_PUBLIC_KEY)")
	// TransactionCmd.PersistentFlags().StringVarP(&clientDo.AddrFlag, "addr", "", defaultAddress(), "specify the account address (for which the public key can be found at monax-keys) (default respects $BURROW_CLIENT_ADDRESS)")
	// TransactionCmd.PersistentFlags().StringVarP(&clientDo.ChainidFlag, "chain-id", "", defaultChainId(), "specify the chainID (default respects $CHAIN_ID)")
	// TransactionCmd.PersistentFlags().StringVarP(&clientDo.NonceFlag, "sequence", "", "", "specify the sequence to use for the transaction (should equal the sender account's sequence + 1)")

	// // TransactionCmd.PersistentFlags().BoolVarP(&clientDo.SignFlag, "sign", "s", false, "sign the transaction using the monax-keys daemon")
	// TransactionCmd.PersistentFlags().BoolVarP(&clientDo.BroadcastFlag, "broadcast", "b", true, "broadcast the transaction to the blockchain")
	// TransactionCmd.PersistentFlags().BoolVarP(&clientDo.WaitFlag, "wait", "w", false, "wait for the transaction to be committed in a block")

	return statusCmd
}
