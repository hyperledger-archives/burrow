package core

import (
	ctypes "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	sm "github.com/eris-ltd/eris-db/manager/eris-mint/state"
	// "github.com/tendermint/tendermint/types"

	// dbm "github.com/tendermint/go-db"
)
// TODO:[ben] remove commented code
//-----------------------------------------------------------------------------

// cache the genesis state
var genesisState *sm.State

// func Status() (*ctypes.ResultStatus, error) {
// 	db := dbm.NewMemDB()
// 	if genesisState == nil {
// 		genesisState = sm.MakeGenesisState(db, genDoc)
// 	}
// 	genesisHash := genesisState.Hash()
// 	latestHeight := blockStore.Height()
// 	var (
// 		latestBlockMeta *types.BlockMeta
// 		latestBlockHash []byte
// 		latestBlockTime int64
// 	)
// 	if latestHeight != 0 {
// 		latestBlockMeta = blockStore.LoadBlockMeta(latestHeight)
// 		latestBlockHash = latestBlockMeta.Hash
// 		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
// 	}
//
// 	return &ctypes.ResultStatus{
// 		NodeInfo:          p2pSwitch.NodeInfo(),
// 		GenesisHash:       genesisHash,
// 		PubKey:            privValidator.PubKey,
// 		LatestBlockHash:   latestBlockHash,
// 		LatestBlockHeight: latestHeight,
// 		LatestBlockTime:   latestBlockTime}, nil
// }

//-----------------------------------------------------------------------------

// func NetInfo() (*ctypes.ResultNetInfo, error) {
// 	listening := p2pSwitch.IsListening()
// 	listeners := []string{}
// 	for _, listener := range p2pSwitch.Listeners() {
// 		listeners = append(listeners, listener.String())
// 	}
// 	peers := []ctypes.Peer{}
// 	for _, peer := range p2pSwitch.Peers().List() {
// 		peers = append(peers, ctypes.Peer{
// 			NodeInfo:   *peer.NodeInfo,
// 			IsOutbound: peer.IsOutbound(),
// 		})
// 	}
// 	return &ctypes.ResultNetInfo{
// 		Listening: listening,
// 		Listeners: listeners,
// 		Peers:     peers,
// 	}, nil
// }

//-----------------------------------------------------------------------------

func Genesis() (*ctypes.ResultGenesis, error) {
	return &ctypes.ResultGenesis{genDoc}, nil
}
