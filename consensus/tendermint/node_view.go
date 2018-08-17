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
	tmNode             *Node
	validatorPublicKey crypto.PublicKey
	txDecoder          txs.Decoder
	runID              simpleuuid.UUID
}

func NewNodeView(tmNode *Node, txDecoder txs.Decoder, runID simpleuuid.UUID) (*NodeView, error) {
	publicKey, err := crypto.PublicKeyFromTendermintPubKey(tmNode.PrivValidator().GetPubKey())
	if err != nil {
		return nil, err
	}
	return &NodeView{
		validatorPublicKey: publicKey,
		tmNode:             tmNode,
		txDecoder:          txDecoder,
		runID:              runID,
	}, nil
}

func (nv *NodeView) ValidatorPublicKey() crypto.PublicKey {
	return nv.validatorPublicKey
}

func (nv *NodeView) NodeInfo() *NodeInfo {
	return NewNodeInfo(nv.tmNode.NodeInfo())
}

func (nv *NodeView) IsListening() bool {
	return nv.tmNode.Switch().IsListening()
}

func (nv *NodeView) IsFastSyncing() bool {
	return nv.tmNode.ConsensusReactor().FastSync()
}

func (nv *NodeView) Listeners() []p2p.Listener {
	return nv.tmNode.Switch().Listeners()
}

func (nv *NodeView) Peers() p2p.IPeerSet {
	return nv.tmNode.Switch().Peers()
}

func (nv *NodeView) BlockStore() state.BlockStoreRPC {
	return nv.tmNode.BlockStore()
}

func (nv *NodeView) RunID() simpleuuid.UUID {
	return nv.runID
}

// Pass -1 to get all available transactions
func (nv *NodeView) MempoolTransactions(maxTxs int) ([]*txs.Envelope, error) {
	var transactions []*txs.Envelope
	for _, txBytes := range nv.tmNode.MempoolReactor().Mempool.Reap(maxTxs) {
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
