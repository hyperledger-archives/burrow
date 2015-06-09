package pipe

import (
	"github.com/tendermint/tendermint/account"
	csus "github.com/tendermint/tendermint/consensus"
	ctypes "github.com/tendermint/tendermint/consensus/types"
	sm "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/types"
)

const (
	MaxUint8 = ^uint8(0) 
	MinUint8 = 0 
	MaxInt8 = int8(MaxUint8 >> 1) 
	MinInt8 = -MaxInt8 - 1
	MaxUint16 = ^uint16(0) 
	MinUint16 = 0 
	MaxInt16 = int16(MaxUint16 >> 1) 
	MinInt16 = -MaxInt16 - 1
	MaxUint = ^uint(0) 
	MinUint = 0 
	MaxInt = int(MaxUint >> 1) 
	MinInt = -MaxInt - 1
	MaxUint64 = ^uint64(0) 
	MinUint64 = 0 
	MaxInt64 = int64(MaxUint64 >> 1) 
	MinInt64 = -MaxInt64 - 1
)

type (

	// *********************************** Address ***********************************

	// Accounts
	AccountList struct {
		Accounts    []*account.Account `json:"accounts"`
	}

	// A contract account storage item.
	StorageItem struct {
		Key   []byte `json:"key"`
		Value []byte `json:"value"`
	}

	// Account storage
	Storage struct {
		StorageRoot  []byte         `json:"storage_root"`
		StorageItems []*StorageItem `json:"storage_items"`
	}

	// *********************************** Blockchain ***********************************

	// BlockchainInfo
	BlockchainInfo struct {
		ChainId           string           `json:"chain_id"`
		GenesisHash       []byte           `json:"genesis_hash"`
		LatestBlockHeight uint             `json:"latest_block_height"`
		LatestBlock       *types.BlockMeta `json:"latest_block"`
	}

	// Genesis hash
	GenesisHash struct {
		Hash []byte `json:"hash"`
	}

	// Get the latest
	LatestBlockHeight struct {
		Height uint `json:"height"`
	}

	ChainId struct {
		ChainId string `json:"chain_id"`
	}

	// GetBlocks
	Blocks struct {
		MinHeight  uint               `json:"min_height"`
		MaxHeight  uint               `json:"max_height"`
		BlockMetas []*types.BlockMeta `json:"block_metas"`
	}

	// *********************************** Consensus ***********************************

	// ConsensusState
	ConsensusState struct {
		Height     uint             `json:"height"`
		Round      uint             `json:"round"`
		Step       uint8            `json:"step"`
		StartTime  string           `json:"start_time"`
		CommitTime string           `json:"commit_time"`
		Validators []*sm.Validator  `json:"validators"`
		Proposal   *ctypes.Proposal `json:"proposal"`
	}

	// Validators
	ValidatorList struct {
		BlockHeight         uint            `json:"block_height"`
		BondedValidators    []*sm.Validator `json:"bonded_validators"`
		UnbondingValidators []*sm.Validator `json:"unbonding_validators"`
	}
	
	// *********************************** Events ***********************************

	// EventSubscribe
	EventSub struct {
		SubId string `json:"sub_id"`
	}

	// EventUnsubscribe
	EventUnsub struct {
		Result bool `json:"result"`
	}

	// EventPoll
	PollResponse struct {
		Events []interface{} `json:"events"`
	}

	// *********************************** Network ***********************************

	// NetworkInfo
	NetworkInfo struct {
		Moniker   string   `json:"moniker"`
		Listening bool     `json:"listening"`
		Listeners []string `json:"listeners"`
		Peers     []*Peer  `json:"peers"`
	}

	Moniker struct {
		Moniker string `json:"moniker"`
	}

	Listening struct {
		Listening bool `json:"listening"`
	}

	Listeners struct {
		Listeners []string `json:"listeners"`
	}

	// used in Peers and BlockchainInfo
	Peer struct {
		nodeInfo   *types.NodeInfo `json:"node_info"`
		IsOutbound bool            `json:"is_outbound"`
	}

	// *********************************** Transactions ***********************************

	// Call or CallCode
	Call struct {
		Return  string `json:"return"`
		GasUsed uint64 `json:"gas_used"`
		// TODO ...
	}

	// UnconfirmedTxs
	UnconfirmedTxs struct {
		Txs []types.Tx `json:"txs"`
	}

	// BroadcastTx or Transact
	Receipt struct {
		TxHash          []byte `json:"tx_hash"`
		CreatesContract uint8  `json:"creates_contract"`
		ContractAddr    []byte `json:"contract_addr"`
	}
	
)

func FromRoundState(rs *csus.RoundState) *ConsensusState {
	cs := &ConsensusState{
		CommitTime: rs.CommitTime.String(),
		Height:     rs.Height,
		Proposal:   rs.Proposal,
		Round:      rs.Round,
		StartTime:  rs.StartTime.String(),
		Step:       uint8(rs.Step),
		Validators: rs.Validators.Validators,
	}
	return cs
}
