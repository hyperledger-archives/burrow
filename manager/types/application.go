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

package types

import (
	// TODO: [ben] this is currently only used for abci result type; but should
	// be removed as abci dependencies shouldn't feature in the application
	// manager
	abci_types "github.com/tendermint/abci/types"
)

// NOTE: [ben] this interface is likely to be changed.  Currently it is taken
// from the tendermint socket protocol application interface;
// but for the needs of eris-rt and generalisation improvements will be made.

// Application interface applies transactions to the state.
type Application interface {

	// Info returns application information as a string
	// NOTE: [ben] likely to move
	Info() (info abci_types.ResponseInfo)

	// Set application option (e.g. mode=mempool, mode=consensus)
	// NOTE: [ben] taken from tendermint, but it is unclear what the use is,
	// specifically, when will tendermint call this over abci ?
	SetOption(key string, value string) (log string)

	// Append transaction applies a transaction to the state regardless of
	// whether the transaction is valid or not.
	// Currently AppendTx is taken from abci, and returns a result.
	// This will be altered, as AppendTransaction needs to more strongly reflect
	// the theoretical logic:
	//   Append(StateN, Transaction) = StateN+1
	// here invalid transactions are allowed, but should act as the identity on
	// the state:
	//   Append(StateN, InvalidTransaction) = StateN
	// TODO: implementation notes:
	// 1. at this point the transaction should already be strongly typed
	// 2.
	DeliverTx(tx []byte) abci_types.Result

	// Check Transaction validates a transaction before being allowed into the
	// consensus' engine memory pool.  This is the original defintion and
	// intention as taken from abci, but should be remapped to the more
	// general concept of basic, cheap verification;
	// Check Transaction does not alter the state, but does require an immutable
	// copy of the state. In particular there is no consensus on ordering yet.
	// TODO: implementation notes:
	// 1. at this point the transaction should already be strongly typed
	// 2.
	CheckTx(tx []byte) abci_types.Result

	// Commit returns the root hash of the current application state
	// NOTE: [ben] Because the concept of the block has been erased here
	// the commit root hash is a fully implict stateful function;
	// the opposit the principle of explicit stateless functions.
	// This will be amended when we introduce the concept of (streaming)
	// blocks in the pipe.
	Commit() abci_types.Result

	// Query for state.  This query request is not passed over the p2p network
	// and is called from Tendermint rpc directly up to the application.
	// NOTE: [ben] Eris-DB will give preference to queries from the local client
	// directly over the Eris-DB rpc.
	// We will support this for Tendermint compatibility.
	Query(query []byte) abci_types.Result
}

// Tendermint has a separate interface for reintroduction of blocks
type BlockchainAware interface {

	// Initialise the blockchain
	// validators: genesis validators from tendermint core
	InitChain(validators []*abci_types.Validator)

	// Signals the beginning of a block;
	// NOTE: [ben] currently not supported by tendermint
	BeginBlock(height uint64)

	// Signals the end of a blockchain
	// validators: changed validators from app to Tendermint
	// NOTE: [ben] currently not supported by tendermint
	// not yet well defined what the change set contains.
	EndBlock(height uint64) (validators []*abci_types.Validator)
}
