package mock

import (
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	"github.com/tendermint/tendermint/account"
	"github.com/tendermint/tendermint/types"
)

// TODO add from js as soon as this serialization stuff is sorted out.
var mockDataJson = `

`

// PipeData is passed into the mock pipe implementation. It provides a
// return value for each of the rpc functions.
type MockData struct {
	PrivAccount       *account.PrivAccount  `json:"priv_account"`
	Account           *account.Account      `json:"account"`
	Accounts          *ep.AccountList       `json:"accounts"`
	Storage           *ep.Storage           `json:"storage"`
	StorageAt         *ep.StorageItem       `json:"storage_at"`
	BlockchainInfo    *ep.BlockchainInfo    `json:"blockchain_status"`
	GenesisHash       *ep.GenesisHash       `json:"genesis_hash"`
	LatestBlockHeight *ep.LatestBlockHeight `json:"latest_block_height"`
	LatestBlock       *types.Block          `json:"latest_block"`
	Blocks            *ep.Blocks            `json:"blocks"`
	Block             *types.Block          `json:"block"`
	ConsensusState    *ep.ConsensusState    `json:"consensus_state"`
	Validators        *ep.ValidatorList     `json:"validators"`
	EventSub          *ep.EventSub          `json:"event_sub"`
	EventUnSub        *ep.EventUnsub        `json:"event_unSub"`
	NetworkInfo       *ep.NetworkInfo       `json:"network_info"`
	Moniker           *ep.Moniker           `json:"moniker"`
	ChainId           *ep.ChainId           `json:"chain_id"`
	Listening         *ep.Listening         `json:"listening"`
	Listeners         *ep.Listeners         `json:"listening"`
	Peers             []*ep.Peer            `json:"peers"`
	Peer              *ep.Peer              `json:"peer"`
	Call              *ep.Call              `json:"call"`
	CallCode          *ep.Call              `json:"call_code"`
	BroadcastTx       *ep.Receipt           `json:"broadcast_tx"`
	UnconfirmedTxs    *ep.UnconfirmedTxs    `json:"unconfirmed_txs"`
	SignTx            *types.CallTx         `json:"sign_tx"`
}

func NewDefaultMockData() *MockData {

	/*
		acc := &account.Account{Address: []byte("0000000000000000000000000000000000000000")}
		acc.Code = []byte{}
		acc.StorageRoot = []byte{}
		accs := make([]*account.Account, 1)
		accs[0] = acc
		accounts := &ep.AccountList{accs}
		storage := &ep.Storage{}
		storageAt := &ep.StorageItem{}

		genesisHash := []byte{0}
		latestBlockHeight := uint(1)
		latestBlock := &ep.Block{}
		chainId := "mock_chain"
		blockchainInfo := &ep.BlockchainInfo{}
		blocks := &ep.Blocks{}
		block := &ep.Block{}

		consensusState := &ep.ConsensusState{}
		validators := &ep.ValidatorList{}

		eventSub := true
		eventUnSub := true
		moniker := "mock_moniker"
		listening := true
		listeners := []string{}
		peer := &ep.Peer{}
		peers := []*ep.Peer{peer}
		networkInfo := &ep.NetworkInfo{moniker, listening, listeners, peers}

		call := &ep.Call{}
		callCode := &ep.Call{}
		broadcastTx := &ep.Receipt{}
		unconfirmedTxs := &ep.UnconfirmedTxs{}
		signTx := &types.CallTx{}
		privAccount := &account.PrivAccount{}

		return &MockData{
			Account:           acc,
			Accounts:          accounts,
			Storage:           storage,
			StorageAt:         storageAt,
			BlockchainInfo:    blockchainInfo,
			GenesisHash:       genesisHash,
			LatestBlockHeight: latestBlockHeight,
			LatestBlock:       latestBlock,
			Blocks:            blocks,
			Block:             block,
			ConsensusState:    consensusState,
			Validators:        validators,
			EventSub:          eventSub,
			EventUnSub:        eventUnSub,
			NetworkInfo:       networkInfo,
			Moniker:           moniker,
			ChainId:           chainId,
			Listening:         listening,
			Listeners:         listeners,
			Peers:             peers,
			Peer:              peer,
			Call:              call,
			CallCode:          callCode,
			BroadcastTx:       broadcastTx,
			UnconfirmedTxs:    unconfirmedTxs,
			SignTx:            signTx,
			PrivAccount:       privAccount,
		}
	*/
	return &MockData{}
}
