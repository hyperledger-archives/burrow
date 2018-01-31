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
	ctypes "github.com/tendermint/tendermint/consensus/types"
	"github.com/tendermint/tendermint/p2p"
	tm_types "github.com/tendermint/tendermint/types"
)

type ResultGetStorage struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
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

type ResultListBlocks struct {
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
	PubKey            acm.PublicKey `json:"pub_key"`
	LatestBlockHash   []byte        `json:"latest_block_hash"`
	LatestBlockHeight uint64        `json:"latest_block_height"`
	LatestBlockTime   int64         `json:"latest_block_time"` // nano
	NodeVersion       string        `json:"node_version"`      // nano
}

type ResultChainId struct {
	ChainName   string `json:"chain_name"`
	ChainId     string `json:"chain_id"`
	GenesisHash []byte `json:"genesis_hash"`
}

type ResultSubscribe struct {
	EventID        string `json:"event"`
	SubscriptionID string `json:"subscription_id"`
}

type ResultUnsubscribe struct {
	SubscriptionID string `json:"subscription_id"`
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

// Type byte helper
var nextByte byte = 1

func biota() (b byte) {
	b = nextByte
	nextByte++
	return
}
