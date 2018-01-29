package query

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/consensus"
	ctypes "github.com/tendermint/tendermint/consensus/types"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/types"
)

// You're like the interface I never had
type NodeView interface {
	// PrivValidator public key
	PrivValidatorPublicKey() acm.PublicKey
	// NodeInfo for this node broadcast to other nodes (including ephemeral STS ED25519 public key)
	NodeInfo() *p2p.NodeInfo
	// Whether the Tendermint node is listening
	IsListening() bool
	// Current listeners
	Listeners() []p2p.Listener
	// Known Tendermint peers
	Peers() p2p.IPeerSet
	// Read-only BlockStore
	BlockStore() types.BlockStoreRPC
	// Get the currently unconfirmed but not known to be invalid transactions from the Node's mempool
	MempoolTransactions(maxTxs int) ([]txs.Tx, error)
	// Get the validator's consensus RoundState
	RoundState() *ctypes.RoundState
	// Get the validator's peer's consensus RoundState
	PeerRoundStates() ([]*ctypes.PeerRoundState, error)
}

type nodeView struct {
	tmNode    *node.Node
	txDecoder txs.Decoder
}

func NewNodeView(tmNode *node.Node, txDecoder txs.Decoder) NodeView {
	return &nodeView{
		tmNode:    tmNode,
		txDecoder: txDecoder,
	}
}

func (nv *nodeView) PrivValidatorPublicKey() acm.PublicKey {
	return acm.PublicKeyFromGoCryptoPubKey(nv.tmNode.PrivValidator().GetPubKey())
}

func (nv *nodeView) NodeInfo() *p2p.NodeInfo {
	return nv.tmNode.NodeInfo()
}

func (nv *nodeView) IsListening() bool {
	return nv.tmNode.Switch().IsListening()
}

func (nv *nodeView) Listeners() []p2p.Listener {
	return nv.tmNode.Switch().Listeners()
}

func (nv *nodeView) Peers() p2p.IPeerSet {
	return nv.tmNode.Switch().Peers()
}

func (nv *nodeView) BlockStore() types.BlockStoreRPC {
	return nv.tmNode.BlockStore()
}

// Pass -1 to get all available transactions
func (nv *nodeView) MempoolTransactions(maxTxs int) ([]txs.Tx, error) {
	var transactions []txs.Tx
	for _, txBytes := range nv.tmNode.MempoolReactor().Mempool.Reap(maxTxs) {
		tx, err := nv.txDecoder.DecodeTx(txBytes)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

func (nv *nodeView) RoundState() *ctypes.RoundState {
	return nv.tmNode.ConsensusState().GetRoundState()
}

func (nv *nodeView) PeerRoundStates() ([]*ctypes.PeerRoundState, error) {
	peers := nv.tmNode.Switch().Peers().List()
	peerRoundStates := make([]*ctypes.PeerRoundState, len(peers))
	for i, peer := range peers {
		peerState, ok := peer.Get(types.PeerStateKey).(*consensus.PeerState)
		if !ok {
			return nil, fmt.Errorf("could not get PeerState for peer: %s", peer)
		}
		peerRoundStates[i] = peerState.GetRoundState()
	}
	return peerRoundStates, nil
}
