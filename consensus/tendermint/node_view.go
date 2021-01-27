package tendermint

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/txs"
	"github.com/streadway/simpleuuid"
	"github.com/tendermint/tendermint/consensus"
	ctypes "github.com/tendermint/tendermint/consensus/types"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
)

type NodeView struct {
	tmNode    *Node
	publicKey crypto.PublicKey
	txDecoder txs.Decoder
	runID     simpleuuid.UUID
}

func NewNodeView(tmNode *Node, txDecoder txs.Decoder, runID simpleuuid.UUID) (*NodeView, error) {
	pk, err := tmNode.PrivValidator().GetPubKey()
	if err != nil {
		return nil, err
	}
	publicKey, err := crypto.PublicKeyFromTendermintPubKey(pk)
	if err != nil {
		return nil, err
	}
	tmNode.BlockStore()
	return &NodeView{
		tmNode:    tmNode,
		publicKey: publicKey,
		txDecoder: txDecoder,
		runID:     runID,
	}, nil
}

func (nv *NodeView) ValidatorPublicKey() crypto.PublicKey {
	if nv == nil {
		return crypto.PublicKey{}
	}
	return nv.publicKey
}

func (nv *NodeView) ValidatorAddress() crypto.Address {
	if nv == nil {
		return crypto.Address{}
	}
	return nv.publicKey.GetAddress()
}

func (nv *NodeView) NodeInfo() *NodeInfo {
	if nv == nil {
		return nil
	}
	ni, ok := nv.tmNode.NodeInfo().(p2p.DefaultNodeInfo)
	if ok {
		return NewNodeInfo(ni)
	}
	return &NodeInfo{}
}

func (nv *NodeView) IsSyncing() bool {
	if nv == nil {
		return true
	}
	return nv.tmNode.ConsensusReactor().WaitSync()
}

func (nv *NodeView) Peers() p2p.IPeerSet {
	return nv.tmNode.Switch().Peers()
}

func (nv *NodeView) BlockStore() state.BlockStore {
	return nv.tmNode.BlockStore()
}

func (nv *NodeView) RunID() simpleuuid.UUID {
	if nv == nil {
		return []byte("00000000-0000-0000-0000-000000000000")
	}
	return nv.runID
}

// Pass -1 to get all available transactions
func (nv *NodeView) MempoolTransactions(maxTxs int) ([]*txs.Envelope, error) {
	var transactions []*txs.Envelope
	for _, txBytes := range nv.tmNode.Mempool().ReapMaxTxs(maxTxs) {
		txEnv, err := nv.txDecoder.DecodeTx(txBytes)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, txEnv)
	}
	return transactions, nil
}

func (nv *NodeView) RoundState() *ctypes.RoundState {
	return nv.tmNode.ConsensusState().GetRoundState()
}

func (nv *NodeView) RoundStateJSON() ([]byte, error) {
	return nv.tmNode.ConsensusState().GetRoundStateJSON()
}

func (nv *NodeView) PeerRoundStates() ([]*ctypes.PeerRoundState, error) {
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
