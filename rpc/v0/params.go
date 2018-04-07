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
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/rpc/filters"
	"github.com/hyperledger/burrow/txs"
)

type (
	// Used to send an address. The address should be hex and properly formatted.
	AddressParam struct {
		Address []byte `json:"address"`
	}

	// Used to send an address
	FilterListParam struct {
		Filters []*filters.FilterData `json:"filters"`
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
		Height uint64 `json:"height"`
	}

	BlocksParam struct {
		MinHeight uint64 `json:"min_height"`
		MaxHeight uint64 `json:"max_height"`
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
		Tx           *txs.CallTx                   `json:"tx"`
		PrivAccounts []*acm.ConcretePrivateAccount `json:"priv_accounts"`
	}

	// Used when sending a transaction to be created and signed on the server
	// (using the private key). This only uses the standard key type for now.
	TransactParam struct {
		PrivKey      []byte `json:"priv_key"`
		InputAddress []byte `json:"input_account"`
		Data         []byte `json:"data"`
		Address      []byte `json:"address"`
		Fee          uint64 `json:"fee"`
		GasLimit     uint64 `json:"gas_limit"`
	}

	// Used when sending a 'Send' transaction.
	SendParam struct {
		PrivKey      []byte `json:"priv_key"`
		InputAddress []byte `json:"input_account"`
		ToAddress    []byte `json:"to_address"`
		Amount       uint64 `json:"amount"`
	}

	NameRegEntryParam struct {
		Name string `json:"name"`
	}

	// Used when sending a namereg transaction to be created and signed on the server
	// (using the private key). This only uses the standard key type for now.
	TransactNameRegParam struct {
		PrivKey      []byte `json:"priv_key"`
		InputAddress []byte `json:"input_account"`
		Name         string `json:"name"`
		Data         string `json:"data"`
		Fee          uint64 `json:"fee"`
		Amount       uint64 `json:"amount"`
	}
)
