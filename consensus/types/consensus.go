package types

import (
	"github.com/eris-ltd/eris-db/event"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-p2p"
	tmsp_types "github.com/tendermint/tmsp/types"
)

type Consensus interface {
	// Peer-2-Peer
	IsListening() bool
	Listeners() []p2p.Listener
	NodeInfo() *p2p.NodeInfo
	Peers() []Peer

	// Private Validator
	PublicValidatorKey() crypto.PubKey

	// Memory pool
	BroadcastTransaction(transaction []byte,
		callback func(*tmsp_types.Response)) error

	// Events
	// For consensus events like NewBlock
	Events() event.EventEmitter
}
