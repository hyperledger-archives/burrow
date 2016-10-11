package types

import (
	"github.com/eris-ltd/eris-db/event"
	"github.com/eris-ltd/eris-db/txs"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-p2p"
	tmsp_types "github.com/tendermint/tmsp/types"
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
		callback func(*tmsp_types.Response)) error

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
}
