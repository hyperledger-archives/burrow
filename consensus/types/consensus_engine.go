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
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/txs"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-p2p"
)

type ConsensusEngine interface {
	// Peer-2-Peer
	IsListening() bool
	Listeners() []p2p.Listener
	NodeInfo() *p2p.NodeInfo
	Peers() []*Peer

	// Private Validator
	PublicValidatorKey() crypto.PubKey

	// Memory pool
	BroadcastTransaction(transaction []byte,
		callback func(*abci_types.Response)) error

	// Events
	// For consensus events like NewBlock
	Events() event.EventEmitter

	// List pending transactions in the mempool, passing 0 for maxTxs gets an
	// unbounded number of transactions
	ListUnconfirmedTxs(maxTxs int) ([]txs.Tx, error)
	ListValidators() []Validator
	ConsensusState() *ConsensusState
	// TODO: Consider creating a real type for PeerRoundState, but at the looks
	// quite coupled to tendermint
	PeerConsensusStates() map[string]string

	// Allow for graceful shutdown of node. Returns whether the node was stopped.
	Stop() bool
}
