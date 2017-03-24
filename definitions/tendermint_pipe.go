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

package definitions

import (
	"github.com/monax/eris-db/account"
	rpc_tm_types "github.com/monax/eris-db/rpc/tendermint/core/types"
	"github.com/monax/eris-db/txs"
)

// NOTE: [ben] TendermintPipe is the additional pipe to carry over
// the RPC exposed by old Tendermint on port `46657` (eris-db-0.11.4 and before)
// This TendermintPipe interface should be deprecated and work towards a generic
// collection of RPC routes for Eris-DB-1.0.0

type TendermintPipe interface {
	Pipe
	// Events
	// Subscribe attempts to subscribe the listener identified by listenerId to
	// the event named event. The Event result is written to rpcResponseWriter
	// which must be non-blocking
	Subscribe(event string,
		rpcResponseWriter func(result rpc_tm_types.ErisDBResult)) (*rpc_tm_types.ResultSubscribe, error)
	Unsubscribe(subscriptionId string) (*rpc_tm_types.ResultUnsubscribe, error)

	// Net
	Status() (*rpc_tm_types.ResultStatus, error)
	NetInfo() (*rpc_tm_types.ResultNetInfo, error)
	Genesis() (*rpc_tm_types.ResultGenesis, error)
	ChainId() (*rpc_tm_types.ResultChainId, error)

	// Accounts
	GetAccount(address []byte) (*rpc_tm_types.ResultGetAccount, error)
	ListAccounts() (*rpc_tm_types.ResultListAccounts, error)
	GetStorage(address, key []byte) (*rpc_tm_types.ResultGetStorage, error)
	DumpStorage(address []byte) (*rpc_tm_types.ResultDumpStorage, error)

	// Call
	Call(fromAddress, toAddress, data []byte) (*rpc_tm_types.ResultCall, error)
	CallCode(fromAddress, code, data []byte) (*rpc_tm_types.ResultCall, error)

	// TODO: [ben] deprecate as we should not allow unsafe behaviour
	// where a user is allowed to send a private key over the wire,
	// especially unencrypted.
	SignTransaction(tx txs.Tx,
		privAccounts []*account.PrivAccount) (*rpc_tm_types.ResultSignTx,
		error)

	// Name registry
	GetName(name string) (*rpc_tm_types.ResultGetName, error)
	ListNames() (*rpc_tm_types.ResultListNames, error)

	// Memory pool
	BroadcastTxAsync(transaction txs.Tx) (*rpc_tm_types.ResultBroadcastTx, error)
	BroadcastTxSync(transaction txs.Tx) (*rpc_tm_types.ResultBroadcastTx, error)

	// Blockchain
	BlockchainInfo(minHeight, maxHeight, maxBlockLookback int) (*rpc_tm_types.ResultBlockchainInfo, error)
	ListUnconfirmedTxs(maxTxs int) (*rpc_tm_types.ResultListUnconfirmedTxs, error)
	GetBlock(height int) (*rpc_tm_types.ResultGetBlock, error)

	// Consensus
	ListValidators() (*rpc_tm_types.ResultListValidators, error)
	DumpConsensusState() (*rpc_tm_types.ResultDumpConsensusState, error)
}
