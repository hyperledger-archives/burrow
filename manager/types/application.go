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

package types

import (
	// TODO: [ben] this is currently only used for tmsp result type; but should
	// be removed as tmsp dependencies shouldn't feature in the application
	// manager
	tmsp_types "github.com/tendermint/tmsp/types"
)

// NOTE: [ben] this interface is likely to be changed.  Currently it is taken
// from the tendermint socket protocol application interface;
// but for the needs of eris-rt and generalisation improvements will be made.

// Application interface applies transactions to the state.
type Application interface {

	// Info returns application information as a string
	// NOTE: [ben] likely to move
	Info() (info string)

	// Set application option (e.g. mode=mempool, mode=consensus)
	// NOTE: [ben] taken from tendermint, but it is unclear what the use is,
	// specifically, when will tendermint call this over tmsp ?
	SetOption(key string, value string) (log string)

	// Append transaction applies a transaction to the state regardless of
	// whether the transaction is valid or not.
	// Currently AppendTx is taken from tmsp, and returns a result.
	// This will be altered, as AppendTransaction needs to more strongly reflect
	// the theoretical logic:
	//   Append(StateN, Transaction) = StateN+1
	// here invalid transactions are allowed, but should act as the identity on
	// the state:
	//   Append(StateN, InvalidTransaction) = StateN
	// TODO: implementation notes:
	// 1. at this point the transaction should already be strongly typed
	// 2.
	AppendTx(tx []byte) tmsp_types.Result

	// Check Transaction validates a transaction before being allowed into the
	// consensus' engine memory pool.  This is the original defintion and
	// intention as taken from tmsp, but should be remapped to the more
	// general concept of basic, cheap verification;
	// Check Transaction does not alter the state, but does require an immutable
	// copy of the state. In particular there is no consensus on ordering yet.
	// TODO: implementation notes:
	// 1. at this point the transaction should already be strongly typed
	// 2.
	CheckTx(tx []byte) tmsp_types.Result

	// Commit returns the root hash of the current application state
	// NOTE: [ben] Because the concept of the block has been erased here
	// the commit root hash is a fully implict stateful function;
	// the opposit the principle of explicit stateless functions.
	// This will be amended when we introduce the concept of (streaming)
	// blocks in the pipe.
	Commit() tmsp_types.Result

	// Query for state.  This query request is not passed over the p2p network
	// and is called from Tendermint rpc directly up to the application.
	// NOTE: [ben] Eris-DB will give preference to queries from the local client
	// directly over the Eris-DB rpc.
	// We will support this for Tendermint compatibility.
	Query(query []byte) tmsp_types.Result
}

// Tendermint has a separate interface for reintroduction of blocks
type BlockchainAware interface {

	// Initialise the blockchain
	// validators: genesis validators from tendermint core
	InitChain(validators []*tmsp_types.Validator)

	// Signals the beginning of a block;
	// NOTE: [ben] currently not supported by tendermint
	BeginBlock(height uint64)

	// Signals the end of a blockchain
	// validators: changed validators from app to Tendermint
	// NOTE: [ben] currently not supported by tendermint
	// not yet well defined what the change set contains.
	EndBlock(height uint64) (validators []*tmsp_types.Validator)
}
