package erisdb

import (
	"github.com/tendermint/tendermint/account"
	"github.com/tendermint/tendermint/types"
	"github.com/eris-ltd/erisdb/erisdb/pipe"
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
	
	// Used to send a tx. Using a string as placeholder until the tx stuff is sorted out.
	TxParam struct {
		Tx types.Tx `json:"tx"`
	}
	
	// StorageAt
	StorageAtParam struct {
		Address []byte `json:"address"`
		Key     []byte `json:"key"`
	}
	
	// Get a block
	HeightParam struct {
		Height uint `json:"height"`
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
		Tx           types.Tx               `json:"tx"`
		PrivAccounts []*account.PrivAccount `json:"priv_accounts"`
	}
	
	// Used when sending a transaction to be created and signed on the server
	// (using the private key).
	TransactParam struct {
		PrivKey  []byte `json:"priv_key"`
		Data     []byte `json:"data"`
		Address  []byte `json:"address"`
		Fee      uint64 `json:"fee"`
		GasLimit uint64 `json:"gas_limit"`
	}
)
