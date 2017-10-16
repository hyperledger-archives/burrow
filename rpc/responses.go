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

package rpc

import (
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/p2p"
	tm_types "github.com/tendermint/tendermint/types"
)

type ResultGetStorage struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

type ResultCall struct {
	execution.Call `json:"unwrap"`
}

type ResultListAccounts struct {
	BlockHeight uint64        `json:"block_height"`
	Accounts    []acm.Account `json:"accounts"`
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
	LastHeight uint64                `json:"last_height"`
	BlockMetas []*tm_types.BlockMeta `json:"block_metas"`
}

type ResultGetBlock struct {
	BlockMeta *tm_types.BlockMeta `json:"block_meta"`
	Block     *tm_types.Block     `json:"block"`
}

type ResultStatus struct {
	NodeInfo          *p2p.NodeInfo `json:"node_info"`
	GenesisHash       []byte        `json:"genesis_hash"`
	PubKey            crypto.PubKey `json:"pub_key"`
	LatestBlockHash   []byte        `json:"latest_block_hash"`
	LatestBlockHeight uint64        `json:"latest_block_height"`
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

type Peer struct {
	NodeInfo   *p2p.NodeInfo `json:"node_info"`
	IsOutbound bool          `json:"is_outbound"`
}

type ResultNetInfo struct {
	Listening bool     `json:"listening"`
	Listeners []string `json:"listeners"`
	Peers     []*Peer  `json:"peers"`
}

type ResultListValidators struct {
	BlockHeight         uint64          `json:"block_height"`
	BondedValidators    []acm.Validator `json:"bonded_validators"`
	UnbondingValidators []acm.Validator `json:"unbonding_validators"`
}

type ResultDumpConsensusState struct {
	RoundState      *consensus.RoundState       `json:"consensus_state"`
	PeerRoundStates []*consensus.PeerRoundState `json:"peer_round_states"`
}

type ResultListNames struct {
	BlockHeight uint64                    `json:"block_height"`
	Names       []*execution.NameRegEntry `json:"names"`
}

type ResultGeneratePrivateAccount struct {
	PrivAccount *acm.ConcretePrivateAccount `json:"priv_account"`
}

type ResultGetAccount struct {
	Account *acm.ConcreteAccount `json:"account"`
}

type ResultBroadcastTx struct {
	*txs.Receipt `json:"unwrap"`
}

type ResultListUnconfirmedTxs struct {
	N   int      `json:"n_txs"`
	Txs []txs.Tx `json:"txs"`
}

type ResultGetName struct {
	Entry *execution.NameRegEntry `json:"entry"`
}

type ResultGenesis struct {
	Genesis genesis.GenesisDoc `json:"genesis"`
}

type ResultSignTx struct {
	Tx txs.Tx `json:"tx"`
}

type ResultEvent struct {
	Event string             `json:"event"`
	Data  event.AnyEventData `json:"data"`
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
}

func ConcreteTypes() []wire.ConcreteType {
	return []wire.ConcreteType{
		{&ResultGetStorage{}, ResultTypeGetStorage},
		{&execution.Call{}, ResultTypeCall},
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
		{&ResultGeneratePrivateAccount{}, ResultTypeGenPrivAccount},
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
