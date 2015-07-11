package erisdb

import (
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
	"github.com/eris-ltd/eris-db/erisdb/pipe"
)

type (

	// Used to send an address. The address should be hex and properly formatted.
	// TODO enforce.
	AddressParam struct {
		Address []byte `json:"address"`
	}

	// Used to send an address
	AccountsParam struct {
		Filters []*pipe.FilterData `json:"filters"`
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
	BlocksParam struct {
		Filters []*pipe.FilterData `json:"filters"`
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
		Data    []byte `json:"data"`
	}

	// Used when doing code calls
	CallCodeParam struct {
		Code []byte `json:"code"`
		Data []byte `json:"data"`
	}

	// Used when signing a tx. Uses placeholders just like TxParam
	SignTxParam struct {
		Tx           *types.CallTx          `json:"tx"`
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
)
