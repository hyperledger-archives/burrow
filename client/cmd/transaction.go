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
	cobra "github.com/spf13/cobra"
)

var TransactionCmd = &cobra.Command(
	Use:   "tx",
	Short: "eris-client tx formulates and signs a transaction to a chain",
	Long:  `eris-client tx formulates and signs a transaction to a chain`,
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 
		if !strings.HasPrefix(clientDo.nodeAddrFlag, "http://") {
			clientDo.nodeAddrFlag = "http://" + clientDo.nodeAddrFlag
		}
		if !strings.HasSuffix(clientDo.nodeAddrFlag, "/") {
			clientDo.nodeAddrFlag += "/"
		}

		if !strings.HasPrefix(clientDo.signAddrFlag, "http://") {
			ClientDo.signAddrFlag = "http://" + clientDo.signAddrFlag
		}
	},
)

func buildTransactionCommand() {
	// Transaction command has subcommands send, name, call, bond,
	// unbond, rebond, permissions. Dupeout transaction is not accessible through the command line.

	addTransactionPersistentFlags()

	// SendTx
	var sendCmd = &cobra.Command{
		Use:   "send",
		Short: "eris-client tx send --amt <amt> --to <addr>",
		Long:  "eris-client tx send --amt <amt> --to <addr>",
		Run:   cliSend,	
	}
	sendCmd.Flags().StringVarP(&clientDo.amtFlag, "amt", "a", "", "specify an amount")
	sendCmd.Flags().StringVarP(&clientDo.toFlag, "to", "t", "", "specify an address to send to")

	// NameTx
	var nameCmd = &cobra.Command{
		Use:   "name",
		Short: "eris-client tx name --amt <amt> --name <name> --data <data>",
		Long:  "eris-client tx name --amt <amt> --name <name> --data <data>",
		Run:   cliName,
	}
	nameCmd.Flags().StringVarP(&clientDo.amtFlag, "amt", "a", "", "specify an amount")
	nameCmd.Flags().StringVarP(&clientDo.nameFlag, "name", "n", "", "specify a name")
	nameCmd.Flags().StringVarP(&clientDo.dataFlag, "data", "d", "", "specify some data")
	nameCmd.Flags().StringVarP(&clientDo.dataFileFlag, "data-file", "", "", "specify a file with some data")
	nameCmd.Flags().StringVarP(&clientDo.feeFlag, "fee", "f", "", "specify the fee to send")

	// CallTx
	var callCmd = &cobra.Command{
		Use:   "call",
		Short: "eris-client tx call --amt <amt> --fee <fee> --gas <gas> --to <contract addr> --data <data>",
		Long:  "eris-client tx call --amt <amt> --fee <fee> --gas <gas> --to <contract addr> --data <data>",
		Run:   cliCall,
	}
	callCmd.Flags().StringVarP(&clientDo.amtFlag, "amt", "a", "", "specify an amount")
	callCmd.Flags().StringVarP(&clientDo.toFlag, "to", "t", "", "specify an address to send to")
	callCmd.Flags().StringVarP(&clientDo.dataFlag, "data", "d", "", "specify some data")
	callCmd.Flags().StringVarP(&clientDo.feeFlag, "fee", "f", "", "specify the fee to send")
	callCmd.Flags().StringVarP(&clientDo.gasFlag, "gas", "g", "", "specify the gas limit for a CallTx")

	// BondTx
	var bondCmd = &cobra.Command{
		Use:   "bond",
		Short: "eris-client tx bond --pubkey <pubkey> --amt <amt> --unbond-to <address>",
		Long:  "eris-client tx bond --pubkey <pubkey> --amt <amt> --unbond-to <address>",
		Run:   cliBond,
	}
	bondCmd.Flags().StringVarP(&clientDo.amtFlag, "amt", "a", "", "specify an amount")
	bondCmd.Flags().StringVarP(&clientDo.unbondtoFlag, "to", "t", "", "specify an address to unbond to")

	// UnbondTx
	var unbondCmd = &cobra.Command{
		Use:   "unbond",
		Short: "eris-client tx unbond --addr <address> --height <block_height>",
		Long:  "eris-client tx unbond --addr <address> --height <block_height>",
		Run:   cliUnbond,
	}
	unbondCmd.Flags().StringVarP(&clientDo.addrFlag, "addr", "a", "", "specify an address")
	unbondCmd.Flags().StringVarP(&clientDo.heightFlag, "height", "n", "", "specify a height to unbond at")

	// RebondTx
	var rebondCmd = &cobra.Command{
		Use:   "rebond",
		Short: "eris-client tx rebond --addr <address> --height <block_height>",
		Long:  "eris-client tx rebond --addr <address> --height <block_height>",
		Run:   cliRebond,
	}
	rebondCmd.Flags().StringVarP(&clientDo.addrFlag, "addr", "a", "", "specify an address")
	rebondCmd.Flags().StringVarP(&clientDo.heightFlag, "height", "n", "", "specify a height to unbond at")

	// PermissionsTx
	var permissionsCmd = &cobra.Command{
		Use:   "permission",
		Short: "eris-client tx perm <function name> <args ...>",
		Long:  "eris-client tx perm <function name> <args ...>",
		Run:   cliPermissions,
	}
	permissionsCmd.Flags().StringVarP(&clientDo.addrFlag, "addr", "a", "", "specify an address")
	permissionsCmd.Flags().StringVarP(&clientDo.heightFlag, "height", "n", "", "specify a height to unbond at")
}

func addTransactionPersistentFlags() {
	ErisClientCmd.PersistentFlags().StringVarP(&clientDo.signAddrFlag, "sign-addr", "", defaultKeyDaemonAddr, "set eris-keys daemon address (default respects $ERIS_CLIENT_SIGN_ADDRESS)")
	ErisClientCmd.PersistentFlags().StringVarP(&clientDo.nodeAddrFlag, "node-addr", "", DefaultNodeRPCAddr, "set the eris-db node rpc server address (default respects $ERIS_CLIENT_NODE_ADDRESS)")
	ErisClientCmd.PersistentFlags().StringVarP(&clientDo.pubkeyFlag, "pubkey", "", defaultPublicKey, "specify the public key to sign with (defaults to $ERIS_CLIENT_PUBLIC_KEY)")
	ErisClientCmd.PersistentFlags().StringVarP(&clientDo.addrFlag, "addr", "", defaultAddress, "specify the account address (from which the public key can be fetch from eris-keys) (default respects $ERIS_CLIENT_ADDRESS)")
	ErisClientCmd.PersistentFlags().StringVarP(&clientDo.chainidFlag, "chain-id", "", defaultChainID, "specify the chainID (default respects $CHAIN_ID)")
	ErisClientCmd.PersistentFlags().StringVarP(&clientDo.nonceFlag, "nonce", "", "", "specify the nonce to use for the transaction (should equal the sender account's nonce + 1)")

	// ErisClientCmd.PersistentFlags().BoolVarP(&signFlag, "sign", "s", false, "sign the transaction using the eris-keys daemon")
	ErisClientCmd.PersistentFlags().BoolVarP(&clientDo.broadcastFlag, "broadcast", "b", false, "broadcast the transaction to the blockchain")
	ErisClientCmd.PersistentFlags().BoolVarP(&clientDo.waitFlag, "wait", "w", false, "wait for the transaction to be committed in a block")
}

//------------------------------------------------------------------------------
// Defaults

func defaultChainId() string {
	return setDefaultString("CHAIN_ID", "")
}

func defaultKeyDaemonAddress() string {
	return setDefaultString("ERIS_CLIENT_SIGN_ADDRESS", "http://localhost:4767")
}

func defaultNodeRpcAddress() string {
	return setDefaultString("ERIS_CLIENT_NODE_ADDRESS", "http://localhost:46657")
}

func defaultPublicKey() string {
	return setDefaultString("ERIS_CLIENT_PUBLIC_KEY", "")
}

func defaultAddress() string {
	return setDefaultString("ERIS_CLIENT_ADDRESS", "")
}
