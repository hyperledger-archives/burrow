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
	"github.com/tendermint/go-wire/data"
	ctypes "github.com/tendermint/tendermint/consensus/types"
	"github.com/tendermint/tendermint/p2p"
	tm_types "github.com/tendermint/tendermint/types"
)

type Result struct {
	ResultInner `json:"unwrap"`
}

type ResultInner interface {
}

func (res Result) Unwrap() ResultInner {
	return res.ResultInner
}

func (br Result) MarshalJSON() ([]byte, error) {
	return mapper.ToJSON(br.ResultInner)
}

func (br *Result) UnmarshalJSON(data []byte) (err error) {
	parsed, err := mapper.FromJSON(data)
	if err == nil && parsed != nil {
		br.ResultInner = parsed.(ResultInner)
	}
	return err
}

var mapper = data.NewMapper(Result{}).
	// Transact
	RegisterImplementation(&ResultBroadcastTx{}, "result_broadcast_tx", biota()).
	// Events
	RegisterImplementation(&ResultSubscribe{}, "result_subscribe", biota()).
	RegisterImplementation(&ResultUnsubscribe{}, "result_unsubscribe", biota()).
	RegisterImplementation(&ResultEvent{}, "result_event", biota()).
	// Status
	RegisterImplementation(&ResultStatus{}, "result_status", biota()).
	RegisterImplementation(&ResultNetInfo{}, "result_net_info", biota()).
	// Accounts
	RegisterImplementation(&ResultGetAccount{}, "result_get_account", biota()).
	RegisterImplementation(&ResultListAccounts{}, "result_list_account", biota()).
	RegisterImplementation(&ResultGetStorage{}, "result_get_storage", biota()).
	RegisterImplementation(&ResultDumpStorage{}, "result_dump_storage", biota()).
	// Simulated call
	RegisterImplementation(&ResultCall{}, "result_call", biota()).
	// Blockchain
	RegisterImplementation(&ResultGenesis{}, "result_genesis", biota()).
	RegisterImplementation(&ResultChainId{}, "result_chain_id", biota()).
	RegisterImplementation(&ResultBlockchainInfo{}, "result_blockchain_info", biota()).
	RegisterImplementation(&ResultGetBlock{}, "result_get_block", biota()).
	// Consensus
	RegisterImplementation(&ResultListUnconfirmedTxs{}, "result_list_unconfirmed_txs", biota()).
	RegisterImplementation(&ResultListValidators{}, "result_list_validators", biota()).
	RegisterImplementation(&ResultDumpConsensusState{}, "result_dump_consensus_state", biota()).
	RegisterImplementation(&ResultPeers{}, "result_peers", biota()).
	// Names
	RegisterImplementation(&ResultGetName{}, "result_get_name", biota()).
	RegisterImplementation(&ResultListNames{}, "result_list_names", biota()).
	// Private keys and signing
	RegisterImplementation(&ResultSignTx{}, "result_sign_tx", biota()).
	RegisterImplementation(&ResultGeneratePrivateAccount{}, "result_generate_private_account", biota())

type ResultGetStorage struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

func (br Result) ResultGetStorage() *ResultGetStorage {
	if res, ok := br.ResultInner.(*ResultGetStorage); ok {
		return res
	}
	return nil
}

type ResultCall struct {
	*execution.Call `json:"unwrap"`
}

type ResultListAccounts struct {
	BlockHeight uint64                 `json:"block_height"`
	Accounts    []*acm.ConcreteAccount `json:"accounts"`
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
	BlockHeight         uint64                   `json:"block_height"`
	BondedValidators    []*acm.ConcreteValidator `json:"bonded_validators"`
	UnbondingValidators []*acm.ConcreteValidator `json:"unbonding_validators"`
}

type ResultDumpConsensusState struct {
	RoundState      *ctypes.RoundState       `json:"consensus_state"`
	PeerRoundStates []*ctypes.PeerRoundState `json:"peer_round_states"`
}

type ResultPeers struct {
	Peers []*Peer `json:"peers"`
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
	N   int           `json:"n_txs"`
	Txs []txs.Wrapper `json:"txs"`
}

type ResultGetName struct {
	Entry *execution.NameRegEntry `json:"entry"`
}

type ResultGenesis struct {
	Genesis genesis.GenesisDoc `json:"genesis"`
}

type ResultSignTx struct {
	Tx txs.Wrapper `json:"tx"`
}

type ResultEvent struct {
	Event              string `json:"event"`
	event.AnyEventData `json:"data"`
}

func (re ResultEvent) Wrap() Result {
	return Result{
		ResultInner: &re,
	}
}

// Type byte helper
var nextByte byte = 1

func biota() (b byte) {
	b = nextByte
	nextByte++
	return
}
