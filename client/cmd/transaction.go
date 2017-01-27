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
	"strings"

	"github.com/spf13/cobra"

	"github.com/eris-ltd/eris-db/client/methods"
	"github.com/eris-ltd/eris-db/util"
)

func buildTransactionCommand() *cobra.Command {
	// Transaction command has subcommands send, name, call, bond,
	// unbond, rebond, permissions. Dupeout transaction is not accessible through the command line.
	transactionCmd := &cobra.Command{
		Use:   "tx",
		Short: "eris-client tx formulates and signs a transaction to a chain",
		Long:  "eris-client tx formulates and signs a transaction to a chain.",
		Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
	}

	addTransactionPersistentFlags(transactionCmd)

	// SendTx
	sendCmd := &cobra.Command{
		Use:   "send",
		Short: "eris-client tx send --amt <amt> --to <addr>",
		Long:  "eris-client tx send --amt <amt> --to <addr>",
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
		Short: "eris-client tx name --amt <amt> --name <name> --data <data>",
		Long:  "eris-client tx name --amt <amt> --name <name> --data <data>",
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
		Short: "eris-client tx call --amt <amt> --fee <fee> --gas <gas> --to <contract addr> --data <data>",
		Long:  "eris-client tx call --amt <amt> --fee <fee> --gas <gas> --to <contract addr> --data <data>",
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
		Short: "eris-client tx bond --pubkey <pubkey> --amt <amt> --unbond-to <address>",
		Long:  "eris-client tx bond --pubkey <pubkey> --amt <amt> --unbond-to <address>",
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
		Short: "eris-client tx unbond --addr <address> --height <block_height>",
		Long:  "eris-client tx unbond --addr <address> --height <block_height>",
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
		Short: "eris-client tx rebond --addr <address> --height <block_height>",
		Long:  "eris-client tx rebond --addr <address> --height <block_height>",
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
		Short: "eris-client tx perm <function name> <args ...>",
		Long:  "eris-client tx perm <function name> <args ...>",
		Run: func(cmd *cobra.Command, args []string) {
			// transaction.Permsissions(clientDo)
		},
		PreRun: assertParameters,
	}

	transactionCmd.AddCommand(sendCmd, nameCmd, callCmd, bondCmd, unbondCmd, rebondCmd, permissionsCmd)
	return transactionCmd
}

func addTransactionPersistentFlags(transactionCmd *cobra.Command) {
	transactionCmd.PersistentFlags().StringVarP(&clientDo.SignAddrFlag, "sign-addr", "", defaultKeyDaemonAddress(), "set eris-keys daemon address (default respects $ERIS_CLIENT_SIGN_ADDRESS)")
	transactionCmd.PersistentFlags().StringVarP(&clientDo.NodeAddrFlag, "node-addr", "", defaultNodeRpcAddress(), "set the eris-db node rpc server address (default respects $ERIS_CLIENT_NODE_ADDRESS)")
	transactionCmd.PersistentFlags().StringVarP(&clientDo.PubkeyFlag, "pubkey", "", defaultPublicKey(), "specify the public key to sign with (defaults to $ERIS_CLIENT_PUBLIC_KEY)")
	transactionCmd.PersistentFlags().StringVarP(&clientDo.AddrFlag, "addr", "", defaultAddress(), "specify the account address (for which the public key can be found at eris-keys) (default respects $ERIS_CLIENT_ADDRESS)")
	transactionCmd.PersistentFlags().StringVarP(&clientDo.ChainidFlag, "chain-id", "", defaultChainId(), "specify the chainID (default respects $CHAIN_ID)")
	transactionCmd.PersistentFlags().StringVarP(&clientDo.NonceFlag, "nonce", "", "", "specify the nonce to use for the transaction (should equal the sender account's nonce + 1)")

	// transactionCmd.PersistentFlags().BoolVarP(&clientDo.SignFlag, "sign", "s", false, "sign the transaction using the eris-keys daemon")
	transactionCmd.PersistentFlags().BoolVarP(&clientDo.BroadcastFlag, "broadcast", "b", true, "broadcast the transaction to the blockchain")
	transactionCmd.PersistentFlags().BoolVarP(&clientDo.WaitFlag, "wait", "w", true, "wait for the transaction to be committed in a block")
}

//------------------------------------------------------------------------------
// Defaults

func defaultChainId() string {
	return setDefaultString("CHAIN_ID", "")
}

func defaultKeyDaemonAddress() string {
	return setDefaultString("ERIS_CLIENT_SIGN_ADDRESS", "http://127.0.0.1:4767")
}

func defaultNodeRpcAddress() string {
	return setDefaultString("ERIS_CLIENT_NODE_ADDRESS", "tcp://127.0.0.1:46657")
}

func defaultPublicKey() string {
	return setDefaultString("ERIS_CLIENT_PUBLIC_KEY", "")
}

func defaultAddress() string {
	return setDefaultString("ERIS_CLIENT_ADDRESS", "")
}

//------------------------------------------------------------------------------
// Helper functions

func assertParameters(cmd *cobra.Command, args []string) {
	if clientDo.ChainidFlag == "" {
		util.Fatalf(`Please provide a chain id either through the flag --chain-id or environment variable $CHAIN_ID.`)
	}

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
