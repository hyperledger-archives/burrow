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

package v0

import (
	"github.com/monax/eris-db/account"
	event "github.com/monax/eris-db/event"
	"github.com/monax/eris-db/txs"
)

type (

	// Used to send an address. The address should be hex and properly formatted.
	// TODO enforce.
	AddressParam struct {
		Address []byte `json:"address"`
	}

	// Used to send an address
	// TODO deprecate in favor of 'FilterListParam'
	AccountsParam struct {
		Filters []*event.FilterData `json:"filters"`
	}

	// Used to send an address
	FilterListParam struct {
		Filters []*event.FilterData `json:"filters"`
	}

	PrivKeyParam struct {
		PrivKey []byte `json:"priv_key"`
	}

	// StorageAt
	StorageAtParam struct {
		Address []byte `json:"address"`
		Key     []byte `json:"key"`
	}

	// Get a block
	HeightParam struct {
		Height int `json:"height"`
	}

	// Get a series of blocks
	// TODO deprecate in favor of 'FilterListParam'
	BlocksParam struct {
		Filters []*event.FilterData `json:"filters"`
	}

	// Event Id
	EventIdParam struct {
		EventId string `json:"event_id"`
	}

	// Event Id
	SubIdParam struct {
		SubId string `json:"sub_id"`
	}

	PeerParam struct {
		Address string `json:"address"`
	}

	// Used when doing calls
	CallParam struct {
		Address []byte `json:"address"`
		From    []byte `json:"from"`
		Data    []byte `json:"data"`
	}

	// Used when doing code calls
	CallCodeParam struct {
		From []byte `json:"from"`
		Code []byte `json:"code"`
		Data []byte `json:"data"`
	}

	// Used when signing a tx. Uses placeholders just like TxParam
	SignTxParam struct {
		Tx           *txs.CallTx            `json:"tx"`
		PrivAccounts []*account.PrivAccount `json:"priv_accounts"`
	}

	// Used when sending a transaction to be created and signed on the server
	// (using the private key). This only uses the standard key type for now.
	TransactParam struct {
		PrivKey  []byte `json:"priv_key"`
		Data     []byte `json:"data"`
		Address  []byte `json:"address"`
		Fee      int64  `json:"fee"`
		GasLimit int64  `json:"gas_limit"`
	}

	// Used when sending a 'Send' transaction.
	SendParam struct {
		PrivKey   []byte `json:"priv_key"`
		ToAddress []byte `json:"to_address"`
		Amount    int64  `json:"amount"`
	}

	NameRegEntryParam struct {
		Name string `json:"name"`
	}

	// Used when sending a namereg transaction to be created and signed on the server
	// (using the private key). This only uses the standard key type for now.
	TransactNameRegParam struct {
		PrivKey []byte `json:"priv_key"`
		Name    string `json:"name"`
		Data    string `json:"data"`
		Fee     int64  `json:"fee"`
		Amount  int64  `json:"amount"`
	}
)
