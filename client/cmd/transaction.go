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
	"strings"

	"github.com/spf13/cobra"

	"github.com/hyperledger/burrow/client/methods"
	"github.com/hyperledger/burrow/util"
)

func buildTransactionCommand() *cobra.Command {
	// Transaction command has subcommands send, name, call, bond,
	// unbond, rebond, permissions.
	transactionCmd := &cobra.Command{
		Use:   "tx",
		Short: "burrow-client tx formulates and signs a transaction to a chain",
		Long:  "burrow-client tx formulates and signs a transaction to a chain.",
		Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
	}

	addTransactionPersistentFlags(transactionCmd)

	// SendTx
	sendCmd := &cobra.Command{
		Use:   "send",
		Short: "burrow-client tx send --amt <amt> --to <addr>",
		Long:  "burrow-client tx send --amt <amt> --to <addr>",
		Run: func(cmd *cobra.Command, args []string) {
			err := methods.Send(clientDo)
			if err != nil {
				util.Fatalf("Could not complete send: %s", err)
			}
		},
		PreRun: assertParameters,
	}
	sendCmd.Flags().StringVarP(&clientDo.AmtFlag, "amt", "a", "", "specify an amount")
	sendCmd.Flags().StringVarP(&clientDo.ToFlag, "to", "t", "", "specify an address to send to")

	// NameTx
	nameCmd := &cobra.Command{
		Use:   "name",
		Short: "burrow-client tx name --amt <amt> --name <name> --data <data>",
		Long:  "burrow-client tx name --amt <amt> --name <name> --data <data>",
		Run: func(cmd *cobra.Command, args []string) {
			// transaction.Name(clientDo)
		},
		PreRun: assertParameters,
	}
	nameCmd.Flags().StringVarP(&clientDo.AmtFlag, "amt", "a", "", "specify an amount")
	nameCmd.Flags().StringVarP(&clientDo.NameFlag, "name", "n", "", "specify a name")
	nameCmd.Flags().StringVarP(&clientDo.DataFlag, "data", "", "", "specify some data")
	nameCmd.Flags().StringVarP(&clientDo.DataFileFlag, "data-file", "", "", "specify a file with some data")
	nameCmd.Flags().StringVarP(&clientDo.FeeFlag, "fee", "f", "", "specify the fee to send")

	// CallTx
	callCmd := &cobra.Command{
		Use:   "call",
		Short: "burrow-client tx call --amt <amt> --fee <fee> --gas <gas> --to <contract addr> --data <data>",
		Long:  "burrow-client tx call --amt <amt> --fee <fee> --gas <gas> --to <contract addr> --data <data>",
		Run: func(cmd *cobra.Command, args []string) {
			err := methods.Call(clientDo)
			if err != nil {
				util.Fatalf("Could not complete call: %s", err)
			}
		},
		PreRun: assertParameters,
	}
	callCmd.Flags().StringVarP(&clientDo.AmtFlag, "amt", "a", "", "specify an amount")
	callCmd.Flags().StringVarP(&clientDo.ToFlag, "to", "t", "", "specify an address to send to")
	callCmd.Flags().StringVarP(&clientDo.DataFlag, "data", "", "", "specify some data")
	callCmd.Flags().StringVarP(&clientDo.FeeFlag, "fee", "f", "", "specify the fee to send")
	callCmd.Flags().StringVarP(&clientDo.GasFlag, "gas", "g", "", "specify the gas limit for a CallTx")

	// BondTx
	bondCmd := &cobra.Command{
		Use:   "bond",
		Short: "burrow-client tx bond --pubkey <pubkey> --amt <amt> --unbond-to <address>",
		Long:  "burrow-client tx bond --pubkey <pubkey> --amt <amt> --unbond-to <address>",
		Run: func(cmd *cobra.Command, args []string) {
			// transaction.Bond(clientDo)
		},
		PreRun: assertParameters,
	}
	bondCmd.Flags().StringVarP(&clientDo.AmtFlag, "amt", "a", "", "specify an amount")
	bondCmd.Flags().StringVarP(&clientDo.UnbondtoFlag, "to", "t", "", "specify an address to unbond to")

	// UnbondTx
	unbondCmd := &cobra.Command{
		Use:   "unbond",
		Short: "burrow-client tx unbond --addr <address> --height <block_height>",
		Long:  "burrow-client tx unbond --addr <address> --height <block_height>",
		Run: func(cmd *cobra.Command, args []string) {
			// transaction.Unbond(clientDo)
		},
		PreRun: assertParameters,
	}
	unbondCmd.Flags().StringVarP(&clientDo.AddrFlag, "addr", "a", "", "specify an address")
	unbondCmd.Flags().StringVarP(&clientDo.HeightFlag, "height", "n", "", "specify a height to unbond at")

	// RebondTx
	var rebondCmd = &cobra.Command{
		Use:   "rebond",
		Short: "burrow-client tx rebond --addr <address> --height <block_height>",
		Long:  "burrow-client tx rebond --addr <address> --height <block_height>",
		Run: func(cmd *cobra.Command, args []string) {
			// transaction.Rebond(clientDo)
		},
		PreRun: assertParameters,
	}
	rebondCmd.Flags().StringVarP(&clientDo.AddrFlag, "addr", "a", "", "specify an address")
	rebondCmd.Flags().StringVarP(&clientDo.HeightFlag, "height", "n", "", "specify a height to unbond at")

	// PermissionsTx
	permissionsCmd := &cobra.Command{
		Use:   "permission",
		Short: "burrow-client tx perm <function name> <args ...>",
		Long:  "burrow-client tx perm <function name> <args ...>",
		Run: func(cmd *cobra.Command, args []string) {
			// transaction.Permsissions(clientDo)
		},
		PreRun: assertParameters,
	}

	transactionCmd.AddCommand(sendCmd, nameCmd, callCmd, bondCmd, unbondCmd, rebondCmd, permissionsCmd)
	return transactionCmd
}

func addTransactionPersistentFlags(transactionCmd *cobra.Command) {
	transactionCmd.PersistentFlags().StringVarP(&clientDo.SignAddrFlag, "sign-addr", "", defaultKeyDaemonAddress(), "set monax-keys daemon address (default respects $BURROW_CLIENT_SIGN_ADDRESS)")
	transactionCmd.PersistentFlags().StringVarP(&clientDo.NodeAddrFlag, "node-addr", "", defaultNodeRpcAddress(), "set the burrow node rpc server address (default respects $BURROW_CLIENT_NODE_ADDRESS)")
	transactionCmd.PersistentFlags().StringVarP(&clientDo.PubkeyFlag, "pubkey", "", defaultPublicKey(), "specify the public key to sign with (defaults to $BURROW_CLIENT_PUBLIC_KEY)")
	transactionCmd.PersistentFlags().StringVarP(&clientDo.AddrFlag, "addr", "", defaultAddress(), "specify the account address (for which the public key can be found at monax-keys) (default respects $BURROW_CLIENT_ADDRESS)")
	transactionCmd.PersistentFlags().StringVarP(&clientDo.NonceFlag, "sequence", "", "", "specify the sequence to use for the transaction (should equal the sender account's sequence + 1)")

	// transactionCmd.PersistentFlags().BoolVarP(&clientDo.SignFlag, "sign", "s", false, "sign the transaction using the monax-keys daemon")
	transactionCmd.PersistentFlags().BoolVarP(&clientDo.BroadcastFlag, "broadcast", "b", true, "broadcast the transaction to the blockchain")
	transactionCmd.PersistentFlags().BoolVarP(&clientDo.WaitFlag, "wait", "w", true, "wait for the transaction to be committed in a block")
}

//------------------------------------------------------------------------------
// Defaults

func defaultKeyDaemonAddress() string {
	return setDefaultString("BURROW_CLIENT_SIGN_ADDRESS", "http://127.0.0.1:4767")
}

func defaultNodeRpcAddress() string {
	return setDefaultString("BURROW_CLIENT_NODE_ADDRESS", "tcp://127.0.0.1:46657")
}

func defaultPublicKey() string {
	return setDefaultString("BURROW_CLIENT_PUBLIC_KEY", "")
}

func defaultAddress() string {
	return setDefaultString("BURROW_CLIENT_ADDRESS", "")
}

//------------------------------------------------------------------------------
// Helper functions

func assertParameters(cmd *cobra.Command, args []string) {
	if !strings.HasPrefix(clientDo.NodeAddrFlag, "tcp://") &&
		!strings.HasPrefix(clientDo.NodeAddrFlag, "unix://") {
		// TODO: [ben] go-rpc will deprecate reformatting; also it is bad practice to auto-correct for this;
		// TODO: [Silas] I've made this fatal, but I'm inclined to define the default as tcp:// and normalise as with http
		// below
		util.Fatalf(`Please use fully formed listening address for the node, including the tcp:// or unix:// prefix.`)
	}

	if !strings.HasPrefix(clientDo.SignAddrFlag, "http://") {
		// NOTE: [ben] we preserve the auto-correction here as it is a simple http request-response to the key server.
		// TODO: [Silas] we don't have logging here to log that we've done this. I'm inclined to either urls without a scheme
		// and be quiet about it, or to make non-compliance fatal
		clientDo.SignAddrFlag = "http://" + clientDo.SignAddrFlag
	}
}
