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
	abci "github.com/tendermint/abci/types"
)

// NOTE: [ben] this interface is likely to be changed.  Currently it is taken
// from the tendermint socket protocol application interface;

// Application interface applies transactions to the state.
type Application interface {

	// Info returns application information as a string
	// NOTE: [ben] likely to move
	Info() (info abci.ResponseInfo)

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
	DeliverTx(tx []byte) abci.Result

	// Check Transaction validates a transaction before being allowed into the
	// consensus' engine memory pool.  This is the original defintion and
	// intention as taken from abci, but should be remapped to the more
	// general concept of basic, cheap verification;
	// Check Transaction does not alter the state, but does require an immutable
	// copy of the state. In particular there is no consensus on ordering yet.
	// TODO: implementation notes:
	// 1. at this point the transaction should already be strongly typed
	// 2.
	CheckTx(tx []byte) abci.Result

	// Commit returns the root hash of the current application state
	// NOTE: [ben] Because the concept of the block has been erased here
	// the commit root hash is a fully implict stateful function;
	// the opposit the principle of explicit stateless functions.
	// This will be amended when we introduce the concept of (streaming)
	// blocks in the pipe.
	Commit() abci.Result

	// Query for state.  This query request is not passed over the p2p network
	// and is called from Tenderpmint rpc directly up to the application.
	// NOTE: [ben] burrow will give preference to queries from the local client
	// directly over the burrow rpc.
	// We will support this for Tendermint compatibility.
	Query(reqQuery abci.RequestQuery) abci.ResponseQuery

	// Tendermint acbi_types.Application extends our base definition of an
	// Application with a parenthetical (begin/end) streaming block interface

	// Initialise the blockchain
	// When Tendermint initialises the genesis validators from tendermint core
	// are passed in as validators
	InitChain(validators []*abci.Validator)

	// Signals the beginning of communicating a block (all transactions have been
	// closed into the block already
	BeginBlock(hash []byte, header *abci.Header)

	// Signals the end of a blockchain
	// ResponseEndBlock wraps a slice of Validators with the Diff field. A Validator
	// is a public key and a voting power. Returning a Validator within this slice
	// asks Tendermint to set that validator's voting power to the Power provided.
	// Note: although the field is named 'Diff' the intention is that it declares
	// the what the new voting power should be (for validators specified,
	// those omitted are left alone) it is not an relative increment to
	// be added (or subtracted) from voting power.
	EndBlock(height uint64) abci.ResponseEndBlock
}
