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
		Address []byte
	}

	// Used to send an address
	FilterListParam struct {
		Filters []*filters.FilterData
	}

	PrivKeyParam struct {
		PrivKey []byte
	}

	// StorageAt
	StorageAtParam struct {
		Address []byte
		Key     []byte
	}

	// Get a block
	HeightParam struct {
		Height uint64
	}

	BlocksParam struct {
		MinHeight uint64
		MaxHeight uint64
	}

	// Event Id
	EventIdParam struct {
		EventId string
	}

	// Event Id
	SubIdParam struct {
		SubId string
	}

	PeerParam struct {
		Address string
	}

	// Used when doing calls
	CallParam struct {
		Address []byte
		From    []byte
		Data    []byte
	}

	// Used when doing code calls
	CallCodeParam struct {
		From []byte
		Code []byte
		Data []byte
	}

	// Used when signing a tx. Uses placeholders just like TxParam
	SignTxParam struct {
		Tx           *txs.CallTx
		PrivAccounts []*acm.ConcretePrivateAccount
	}

	// Used when sending a transaction to be created and signed on the server
	// (using the private key). This only uses the standard key type for now.
	TransactParam struct {
		PrivKey      []byte
		InputAddress []byte
		Data         []byte
		Address      []byte
		Fee          uint64
		GasLimit     uint64
	}

	// Used when sending a 'Send' transaction.
	SendParam struct {
		PrivKey      []byte
		InputAddress []byte
		ToAddress    []byte
		Amount       uint64
	}

	NameRegEntryParam struct {
		Name string
	}

	// Used when sending a namereg transaction to be created and signed on the server
	// (using the private key). This only uses the standard key type for now.
	TransactNameRegParam struct {
		PrivKey      []byte
		InputAddress []byte
		Name         string
		Data         string
		Fee          uint64
		Amount       uint64
	}
)
