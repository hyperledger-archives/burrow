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
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/tendermint/go-wire/data"
)

// Legacy for JS
var _ = data.NewMapper(struct{ payload.Payload }{}).
	RegisterImplementation(&payload.SendTx{}, "send_tx", byte(payload.TypeSend)).
	RegisterImplementation(&payload.CallTx{}, "call_tx", byte(payload.TypeCall)).
	RegisterImplementation(&payload.NameTx{}, "name_tx", byte(payload.TypeName)).
	RegisterImplementation(&payload.BondTx{}, "bond_tx", byte(payload.TypeBond)).
	RegisterImplementation(&payload.UnbondTx{}, "unbond_tx", byte(payload.TypeUnbond)).
	RegisterImplementation(&payload.PermissionsTx{}, "permissions_tx", byte(payload.TypePermissions))

type (
	// Used to send an address. The address should be hex and properly formatted.
	AddressParam struct {
		Address []byte `json:"address"`
	}

	// Used to send an address
	FilterListParam struct {
		Filters []*filters.FilterData `json:"filters"`
	}

	PrivateKeyParam struct {
		PrivateKey []byte `json:"privateKey"`
	}

	InputAccount struct {
		PrivateKey []byte `json:"privateKey"`
		Address    []byte `json:"address"`
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
		Tx              *payload.CallTx               `json:"tx"`
		PrivateAccounts []*acm.ConcretePrivateAccount `json:"privateAccounts"`
	}

	// Used when sending a transaction to be created and signed on the server
	// (using the private key). This only uses the standard key type for now.
	TransactParam struct {
		InputAccount InputAccount `json:"inputAccount"`
		Data         []byte       `json:"data"`
		Address      []byte       `json:"address"`
		Fee          uint64       `json:"fee"`
		GasLimit     uint64       `json:"gasLimit"`
	}

	// Used when sending a 'Send' transaction.
	SendParam struct {
		InputAccount InputAccount `json:"inputAccount"`
		ToAddress    []byte       `json:"toAddress"`
		Amount       uint64       `json:"amount"`
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
