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
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/rpc/filters"
	"github.com/hyperledger/burrow/txs/payload"
)

type (
	// Used to send an address. The address should be hex and properly formatted.
	AddressParam struct {
		Address binary.HexBytes `json:"address"`
	}

	// Used to send an address
	FilterListParam struct {
		Filters []*filters.FilterData `json:"filters"`
	}

	PrivateKeyParam struct {
		PrivateKey binary.HexBytes `json:"privateKey"`
	}

	InputAccount struct {
		PrivateKey binary.HexBytes `json:"privateKey"`
		Address    binary.HexBytes `json:"address"`
	}

	// StorageAt
	StorageAtParam struct {
		Address binary.HexBytes `json:"address"`
		Key     binary.HexBytes `json:"key"`
	}

	// Get a block
	HeightParam struct {
		Height uint64 `json:"height"`
	}

	BlocksParam struct {
		MinHeight uint64 `json:"minHeight"`
		MaxHeight uint64 `json:"maxHeight"`
	}

	// Event Id
	EventIdParam struct {
		EventId string `json:"eventId"`
	}

	// Event Id
	SubIdParam struct {
		SubId string `json:"subId"`
	}

	PeerParam struct {
		Address string `json:"address"`
	}

	// Used when doing calls
	CallParam struct {
		Address binary.HexBytes `json:"address"`
		From    binary.HexBytes `json:"from"`
		Data    binary.HexBytes `json:"data"`
	}

	// Used when doing code calls
	CallCodeParam struct {
		From binary.HexBytes `json:"from"`
		Code binary.HexBytes `json:"code"`
		Data binary.HexBytes `json:"data"`
	}

	// Used when signing a tx. Uses placeholders just like TxParam
	SignTxParam struct {
		Tx              *payload.CallTx               `json:"tx"`
		PrivateAccounts []*acm.ConcretePrivateAccount `json:"privateAccounts"`
	}

	// Used when sending a transaction to be created and signed on the server
	// (using the private key). This only uses the standard key type for now.
	TransactParam struct {
		InputAccount InputAccount    `json:"inputAccount"`
		Data         binary.HexBytes `json:"data"`
		Address      binary.HexBytes `json:"address"`
		Fee          uint64          `json:"fee"`
		GasLimit     uint64          `json:"gasLimit"`
	}

	// Used when sending a 'Send' transaction.
	SendParam struct {
		InputAccount InputAccount    `json:"inputAccount"`
		ToAddress    binary.HexBytes `json:"toAddress"`
		Amount       uint64          `json:"amount"`
	}

	NameRegEntryParam struct {
		Name string `json:"name"`
	}

	// Used when sending a namereg transaction to be created and signed on the server
	// (using the private key). This only uses the standard key type for now.
	TransactNameRegParam struct {
		InputAccount InputAccount `json:"inputAccount"`
		Name         string       `json:"name"`
		Data         string       `json:"data"`
		Fee          uint64       `json:"fee"`
		Amount       uint64       `json:"amount"`
	}
)
