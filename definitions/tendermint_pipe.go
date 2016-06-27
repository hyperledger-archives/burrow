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

package definitions

import (
	account              "github.com/eris-ltd/eris-db/account"
	rpc_tendermint_types "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	transaction          "github.com/eris-ltd/eris-db/txs"
)

// NOTE: [ben] TendermintPipe is the additional pipe to carry over
// the RPC exposed by old Tendermint on port `46657` (eris-db-0.11.4 and before)
// This TendermintPipe interface should be deprecated and work towards a generic
// collection of RPC routes for Eris-DB-1.0.0

type TendermintPipe interface {
	// Net
	Status() (*rpc_tendermint_types.ResultStatus, error)

	NetInfo() (*rpc_tendermint_types.ResultNetInfo, error)

	Genesis() (*rpc_tendermint_types.ResultGenesis, error)

	// Accounts
	GetAccount(address []byte) (*rpc_tendermint_types.ResultGetAccount,
		error)

	ListAccounts() (*rpc_tendermint_types.ResultListAccounts, error)

	GetStorage(address, key []byte) (*rpc_tendermint_types.ResultGetStorage,
		error)

	DumpStorage(address []byte) (*rpc_tendermint_types.ResultDumpStorage,
		error)

	// Call
	Call(fromAddres, toAddress, data []byte) (*rpc_tendermint_types.ResultCall,
		error)

	CallCode(fromAddress, code, data []byte) (*rpc_tendermint_types.ResultCall,
		error)

	// TODO: [ben] deprecate as we should not allow unsafe behaviour
	// where a user is allowed to send a private key over the wire,
	// especially unencrypted.
	SignTransaction(transaction transaction.Tx,
		privAccounts []*account.PrivAccount) (*rpc_tendermint_types.ResultSignTx,
		error)

	// Name registry
	GetName(name string) (*rpc_tendermint_types.ResultGetName, error)

	ListNames() (*rpc_tendermint_types.ResultListNames, error)

	// Memory pool
	BroadcastTxAsync(transaction transaction.Tx) (*rpc_tendermint_types.ResultBroadcastTx,
		error)

	BroadcastTxSync(transaction transaction.Tx) (*rpc_tendermint_types.ResultBroadcastTx,
		error)
}
