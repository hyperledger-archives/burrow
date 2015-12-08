package pipe

import (
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/go-p2p" // NodeInfo (drop this!)
	csus "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/consensus"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
	"github.com/eris-ltd/eris-db/account"
)

type (

	// *********************************** Address ***********************************

	// Accounts
	AccountList struct {
		Accounts []*account.Account `json:"accounts"`
	}

	// A contract account storage item.
	StorageItem struct {
		Key   []byte `json:"key"`
		Value []byte `json:"value"`
	}

	// Account storage
	Storage struct {
		StorageRoot  []byte        `json:"storage_root"`
		StorageItems []StorageItem `json:"storage_items"`
	}

	// *********************************** Blockchain ***********************************

	// BlockchainInfo
	BlockchainInfo struct {
		ChainId           string           `json:"chain_id"`
		GenesisHash       []byte           `json:"genesis_hash"`
		LatestBlockHeight int              `json:"latest_block_height"`
		LatestBlock       *types.BlockMeta `json:"latest_block"`
	}

	// Genesis hash
	GenesisHash struct {
		Hash []byte `json:"hash"`
	}

	// Get the latest
	LatestBlockHeight struct {
		Height int `json:"height"`
	}

	ChainId struct {
		ChainId string `json:"chain_id"`
	}

	// GetBlocks
	Blocks struct {
		MinHeight  int                `json:"min_height"`
		MaxHeight  int                `json:"max_height"`
		BlockMetas []*types.BlockMeta `json:"block_metas"`
	}

	// *********************************** Consensus ***********************************

	// ConsensusState
	ConsensusState struct {
		Height     int                `json:"height"`
		Round      int                `json:"round"`
		Step       uint8              `json:"step"`
		StartTime  string             `json:"start_time"`
		CommitTime string             `json:"commit_time"`
		Validators []*types.Validator `json:"validators"`
		Proposal   *types.Proposal    `json:"proposal"`
	}

	// Validators
	ValidatorList struct {
		BlockHeight         int                `json:"block_height"`
		BondedValidators    []*types.Validator `json:"bonded_validators"`
		UnbondingValidators []*types.Validator `json:"unbonding_validators"`
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
		ClientVersion string   `json:"client_version"`
		Moniker       string   `json:"moniker"`
		Listening     bool     `json:"listening"`
		Listeners     []string `json:"listeners"`
		Peers         []*Peer  `json:"peers"`
	}

	ClientVersion struct {
		ClientVersion string `json:"client_version"`
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
		nodeInfo   *p2p.NodeInfo `json:"node_info"`
		IsOutbound bool          `json:"is_outbound"`
	}

	// *********************************** Transactions ***********************************

	// Call or CallCode
	Call struct {
		Return  string `json:"return"`
		GasUsed int64  `json:"gas_used"`
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

	TransactionResult struct {
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
