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
	"encoding/json"
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/execution"
	exe_events "github.com/hyperledger/burrow/execution/events"
	evm_events "github.com/hyperledger/burrow/execution/evm/events"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/txs"
	ctypes "github.com/tendermint/tendermint/consensus/types"
	"github.com/tendermint/tendermint/p2p"
	tm_types "github.com/tendermint/tendermint/types"
)

type ResultGetStorage struct {
	Key   []byte
	Value []byte
}

type ResultCall struct {
	execution.Call
}

func (rc ResultCall) MarshalJSON() ([]byte, error) {
	return json.Marshal(rc.Call)
}

func (rc *ResultCall) UnmarshalJSON(data []byte) (err error) {
	return json.Unmarshal(data, &rc.Call)
}

type ResultListAccounts struct {
	BlockHeight uint64
	Accounts    []*acm.ConcreteAccount
}

type ResultDumpStorage struct {
	StorageRoot  []byte
	StorageItems []StorageItem
}

type StorageItem struct {
	Key   []byte
	Value []byte
}

type ResultListBlocks struct {
	LastHeight uint64
	BlockMetas []*tm_types.BlockMeta
}

type ResultGetBlock struct {
	BlockMeta *tm_types.BlockMeta
	Block     *tm_types.Block
}

type ResultStatus struct {
	NodeInfo          p2p.NodeInfo
	GenesisHash       []byte
	PubKey            acm.PublicKey
	LatestBlockHash   []byte
	LatestBlockHeight uint64
	LatestBlockTime   int64
	NodeVersion       string
}

type ResultChainId struct {
	ChainName   string
	ChainId     string
	GenesisHash []byte
}

type ResultSubscribe struct {
	EventID        string
	SubscriptionID string
}

type ResultUnsubscribe struct {
	SubscriptionID string
}

type Peer struct {
	NodeInfo   p2p.NodeInfo
	IsOutbound bool
}

type ResultNetInfo struct {
	Listening bool
	Listeners []string
	Peers     []*Peer
}

type ResultListValidators struct {
	BlockHeight         uint64
	BondedValidators    []*acm.ConcreteValidator
	UnbondingValidators []*acm.ConcreteValidator
}

type ResultDumpConsensusState struct {
	RoundState      *ctypes.RoundState
	PeerRoundStates []*ctypes.PeerRoundState
}

type ResultPeers struct {
	Peers []*Peer
}

type ResultListNames struct {
	BlockHeight uint64
	Names       []*execution.NameRegEntry
}

type ResultGeneratePrivateAccount struct {
	PrivateAccount *acm.ConcretePrivateAccount
}

type ResultGetAccount struct {
	Account *acm.ConcreteAccount
}

type AccountHumanReadable struct {
	Address     acm.Address
	PublicKey   acm.PublicKey
	Sequence    uint64
	Balance     uint64
	Code        []string
	StorageRoot string
	Permissions []string
	Roles       []string
}

type ResultGetAccountHumanReadable struct {
	Account *AccountHumanReadable
}

type ResultBroadcastTx struct {
	txs.Receipt
}

// Unwrap

func (rbt ResultBroadcastTx) MarshalJSON() ([]byte, error) {
	return json.Marshal(rbt.Receipt)
}

func (rbt ResultBroadcastTx) UnmarshalJSON(data []byte) (err error) {
	return json.Unmarshal(data, &rbt.Receipt)
}

type ResultListUnconfirmedTxs struct {
	NumTxs int
	Txs    []txs.Wrapper
}

type ResultGetName struct {
	Entry *execution.NameRegEntry
}

type ResultGenesis struct {
	Genesis genesis.GenesisDoc
}

type ResultSignTx struct {
	Tx txs.Wrapper
}

type ResultEvent struct {
	Event         string
	TMEventData   *tm_types.TMEventData     `json:",omitempty"`
	EventDataTx   *exe_events.EventDataTx   `json:",omitempty"`
	EventDataCall *evm_events.EventDataCall `json:",omitempty"`
	EventDataLog  *evm_events.EventDataLog  `json:",omitempty"`
}

func (resultEvent ResultEvent) EventDataNewBlock() *tm_types.EventDataNewBlock {
	if resultEvent.TMEventData != nil {
		eventData, _ := resultEvent.TMEventData.Unwrap().(tm_types.EventDataNewBlock)
		return &eventData
	}
	return nil
}

// Map any supported event data element to our ResultEvent sum type
func NewResultEvent(event string, eventData interface{}) (*ResultEvent, error) {
	switch ed := eventData.(type) {
	case tm_types.TMEventData:
		return &ResultEvent{
			Event:       event,
			TMEventData: &ed,
		}, nil

	case *exe_events.EventDataTx:
		return &ResultEvent{
			Event:       event,
			EventDataTx: ed,
		}, nil

	case *evm_events.EventDataCall:
		return &ResultEvent{
			Event:         event,
			EventDataCall: ed,
		}, nil

	case *evm_events.EventDataLog:
		return &ResultEvent{
			Event:        event,
			EventDataLog: ed,
		}, nil

	default:
		return nil, fmt.Errorf("could not map event data of type %T to ResultEvent", eventData)
	}
}
