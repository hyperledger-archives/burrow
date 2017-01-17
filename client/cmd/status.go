// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package commands

import (
	"github.com/spf13/cobra"

	"github.com/eris-ltd/eris-db/client/methods"
)


func buildStatusCommand() *cobra.Command{
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "eris-client status returns the current status from a chain.",
		Long: `eris-client status returns the current status from a chain.
`,
		Run: func(cmd *cobra.Command, args []string) {
			methods.Status(clientDo)
		},
	}
	statusCmd.PersistentFlags().StringVarP(&clientDo.NodeAddrFlag, "node-addr", "", defaultNodeRpcAddress(), "set the eris-db node rpc server address (default respects $ERIS_CLIENT_NODE_ADDRESS)")
	// TransactionCmd.PersistentFlags().StringVarP(&clientDo.PubkeyFlag, "pubkey", "", defaultPublicKey(), "specify the public key to sign with (defaults to $ERIS_CLIENT_PUBLIC_KEY)")
	// TransactionCmd.PersistentFlags().StringVarP(&clientDo.AddrFlag, "addr", "", defaultAddress(), "specify the account address (for which the public key can be found at eris-keys) (default respects $ERIS_CLIENT_ADDRESS)")
	// TransactionCmd.PersistentFlags().StringVarP(&clientDo.ChainidFlag, "chain-id", "", defaultChainId(), "specify the chainID (default respects $CHAIN_ID)")
	// TransactionCmd.PersistentFlags().StringVarP(&clientDo.NonceFlag, "nonce", "", "", "specify the nonce to use for the transaction (should equal the sender account's nonce + 1)")

	// // TransactionCmd.PersistentFlags().BoolVarP(&clientDo.SignFlag, "sign", "s", false, "sign the transaction using the eris-keys daemon")
	// TransactionCmd.PersistentFlags().BoolVarP(&clientDo.BroadcastFlag, "broadcast", "b", true, "broadcast the transaction to the blockchain")
	// TransactionCmd.PersistentFlags().BoolVarP(&clientDo.WaitFlag, "wait", "w", false, "wait for the transaction to be committed in a block")

	return statusCmd
}

