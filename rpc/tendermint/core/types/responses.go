// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	acm "github.com/hyperledger/burrow/account"
	core_types "github.com/hyperledger/burrow/core/types"
	genesis "github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/txs"
	tendermint_types "github.com/tendermint/tendermint/types"

	consensus_types "github.com/hyperledger/burrow/consensus/types"
	abcitypes "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-p2p"
	"github.com/tendermint/go-rpc/types"
	"github.com/tendermint/go-wire"
)

type ResultGetStorage struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

type ResultCall struct {
	Return  []byte `json:"return"`
	GasUsed int64  `json:"gas_used"`
	// TODO ...
}

type ResultListAccounts struct {
	BlockHeight int            `json:"block_height"`
	Accounts    []*acm.Account `json:"accounts"`
}

type ResultDumpStorage struct {
	StorageRoot  []byte        `json:"storage_root"`
	StorageItems []StorageItem `json:"storage_items"`
}

type StorageItem struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

type ResultBlockchainInfo struct {
	LastHeight int                           `json:"last_height"`
	BlockMetas []*tendermint_types.BlockMeta `json:"block_metas"`
}

type ResultGetBlock struct {
	BlockMeta *tendermint_types.BlockMeta `json:"block_meta"`
	Block     *tendermint_types.Block     `json:"block"`
}

type ResultStatus struct {
	NodeInfo          *p2p.NodeInfo `json:"node_info"`
	GenesisHash       []byte        `json:"genesis_hash"`
	PubKey            crypto.PubKey `json:"pub_key"`
	LatestBlockHash   []byte        `json:"latest_block_hash"`
	LatestBlockHeight int           `json:"latest_block_height"`
	LatestBlockTime   int64         `json:"latest_block_time"` // nano
}

type ResultChainId struct {
	ChainName   string `json:"chain_name"`
	ChainId     string `json:"chain_id"`
	GenesisHash []byte `json:"genesis_hash"`
}

type ResultSubscribe struct {
	Event          string `json:"event"`
	SubscriptionId string `json:"subscription_id"`
}

type ResultUnsubscribe struct {
	SubscriptionId string `json:"subscription_id"`
}

type ResultNetInfo struct {
	Listening bool                    `json:"listening"`
	Listeners []string                `json:"listeners"`
	Peers     []*consensus_types.Peer `json:"peers"`
}

type ResultListValidators struct {
	BlockHeight         int                         `json:"block_height"`
	BondedValidators    []consensus_types.Validator `json:"bonded_validators"`
	UnbondingValidators []consensus_types.Validator `json:"unbonding_validators"`
}

type ResultDumpConsensusState struct {
	ConsensusState      *consensus_types.ConsensusState `json:"consensus_state"`
	PeerConsensusStates []*ResultPeerConsensusState     `json:"peer_consensus_states"`
}

type ResultPeerConsensusState struct {
	PeerKey            string `json:"peer_key"`
	PeerConsensusState string `json:"peer_consensus_state"`
}

type ResultListNames struct {
	BlockHeight int                        `json:"block_height"`
	Names       []*core_types.NameRegEntry `json:"names"`
}

type ResultGenPrivAccount struct {
	PrivAccount *acm.PrivAccount `json:"priv_account"`
}

type ResultGetAccount struct {
	Account *acm.Account `json:"account"`
}

type ResultBroadcastTx struct {
	Code abcitypes.CodeType `json:"code"`
	Data []byte             `json:"data"`
	Log  string             `json:"log"`
}

type ResultListUnconfirmedTxs struct {
	N   int      `json:"n_txs"`
	Txs []txs.Tx `json:"txs"`
}

type ResultGetName struct {
	Entry *core_types.NameRegEntry `json:"entry"`
}

type ResultGenesis struct {
	Genesis *genesis.GenesisDoc `json:"genesis"`
}

type ResultSignTx struct {
	Tx txs.Tx `json:"tx"`
}

type ResultEvent struct {
	Event string        `json:"event"`
	Data  txs.EventData `json:"data"`
}

//----------------------------------------
// result types

const (
	ResultTypeGetStorage         = byte(0x01)
	ResultTypeCall               = byte(0x02)
	ResultTypeListAccounts       = byte(0x03)
	ResultTypeDumpStorage        = byte(0x04)
	ResultTypeBlockchainInfo     = byte(0x05)
	ResultTypeGetBlock           = byte(0x06)
	ResultTypeStatus             = byte(0x07)
	ResultTypeNetInfo            = byte(0x08)
	ResultTypeListValidators     = byte(0x09)
	ResultTypeDumpConsensusState = byte(0x0A)
	ResultTypeListNames          = byte(0x0B)
	ResultTypeGenPrivAccount     = byte(0x0C)
	ResultTypeGetAccount         = byte(0x0D)
	ResultTypeBroadcastTx        = byte(0x0E)
	ResultTypeListUnconfirmedTxs = byte(0x0F)
	ResultTypeGetName            = byte(0x10)
	ResultTypeGenesis            = byte(0x11)
	ResultTypeSignTx             = byte(0x12)
	ResultTypeEvent              = byte(0x13) // so websockets can respond to rpc functions
	ResultTypeSubscribe          = byte(0x14)
	ResultTypeUnsubscribe        = byte(0x15)
	ResultTypePeerConsensusState = byte(0x16)
	ResultTypeChainId            = byte(0x17)
)

type BurrowResult interface {
	rpctypes.Result
}

func ConcreteTypes() []wire.ConcreteType {
	return []wire.ConcreteType{
		{&ResultGetStorage{}, ResultTypeGetStorage},
		{&ResultCall{}, ResultTypeCall},
		{&ResultListAccounts{}, ResultTypeListAccounts},
		{&ResultDumpStorage{}, ResultTypeDumpStorage},
		{&ResultBlockchainInfo{}, ResultTypeBlockchainInfo},
		{&ResultGetBlock{}, ResultTypeGetBlock},
		{&ResultStatus{}, ResultTypeStatus},
		{&ResultNetInfo{}, ResultTypeNetInfo},
		{&ResultListValidators{}, ResultTypeListValidators},
		{&ResultDumpConsensusState{}, ResultTypeDumpConsensusState},
		{&ResultDumpConsensusState{}, ResultTypePeerConsensusState},
		{&ResultListNames{}, ResultTypeListNames},
		{&ResultGenPrivAccount{}, ResultTypeGenPrivAccount},
		{&ResultGetAccount{}, ResultTypeGetAccount},
		{&ResultBroadcastTx{}, ResultTypeBroadcastTx},
		{&ResultListUnconfirmedTxs{}, ResultTypeListUnconfirmedTxs},
		{&ResultGetName{}, ResultTypeGetName},
		{&ResultGenesis{}, ResultTypeGenesis},
		{&ResultSignTx{}, ResultTypeSignTx},
		{&ResultEvent{}, ResultTypeEvent},
		{&ResultSubscribe{}, ResultTypeSubscribe},
		{&ResultUnsubscribe{}, ResultTypeUnsubscribe},
		{&ResultChainId{}, ResultTypeChainId},
	}
}

var _ = wire.RegisterInterface(struct{ BurrowResult }{}, ConcreteTypes()...)
