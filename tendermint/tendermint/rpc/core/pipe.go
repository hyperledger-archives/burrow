package core

import (
	bc "github.com/eris-ltd/eris-db/tendermint/tendermint/blockchain"
	"github.com/eris-ltd/eris-db/tendermint/tendermint/consensus"
	mempl "github.com/eris-ltd/eris-db/tendermint/tendermint/mempool"
	"github.com/eris-ltd/eris-db/tendermint/tendermint/p2p"
	stypes "github.com/eris-ltd/eris-db/tendermint/tendermint/state/types"
	"github.com/eris-ltd/eris-db/tendermint/tendermint/types"
)

var blockStore *bc.BlockStore
var consensusState *consensus.ConsensusState
var consensusReactor *consensus.ConsensusReactor
var mempoolReactor *mempl.MempoolReactor
var p2pSwitch *p2p.Switch
var privValidator *types.PrivValidator
var genDoc *stypes.GenesisDoc // cache the genesis structure

func SetBlockStore(bs *bc.BlockStore) {
	blockStore = bs
}

func SetConsensusState(cs *consensus.ConsensusState) {
	consensusState = cs
}

func SetConsensusReactor(cr *consensus.ConsensusReactor) {
	consensusReactor = cr
}

func SetMempoolReactor(mr *mempl.MempoolReactor) {
	mempoolReactor = mr
}

func SetSwitch(sw *p2p.Switch) {
	p2pSwitch = sw
}

func SetPrivValidator(pv *types.PrivValidator) {
	privValidator = pv
}

func SetGenDoc(doc *stypes.GenesisDoc) {
	genDoc = doc
}
