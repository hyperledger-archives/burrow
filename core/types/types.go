// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

// TODO: [ben] this is poorly constructed but copied over
// from eris-db/erisdb/pipe/types to make incremental changes and allow
// for a discussion around the proper defintion of the needed types.
package types

import (
	"github.com/tendermint/go-p2p" // NodeInfo (drop this!)
	"github.com/tendermint/tendermint/types"

	account "github.com/eris-ltd/eris-db/account"
)

type (
	// *********************************** Address ***********************************

	// Accounts
	AccountList struct {
		Accounts []*account.Account `json:"accounts"`
	}

	// A contract account storage item.
	StorageItem struct {
		Key   []byte `json:"key"`
		Value []byte `json:"value"`
	}

	// Account storage
	Storage struct {
		StorageRoot  []byte        `json:"storage_root"`
		StorageItems []StorageItem `json:"storage_items"`
	}

	// *********************************** Blockchain ***********************************

	// BlockchainInfo
	BlockchainInfo struct {
		ChainId           string           `json:"chain_id"`
		GenesisHash       []byte           `json:"genesis_hash"`
		LatestBlockHeight int              `json:"latest_block_height"`
		LatestBlock       *types.BlockMeta `json:"latest_block"`
	}

	// Genesis hash
	GenesisHash struct {
		Hash []byte `json:"hash"`
	}

	// Get the latest
	LatestBlockHeight struct {
		Height int `json:"height"`
	}

	ChainId struct {
		ChainId string `json:"chain_id"`
	}

	// GetBlocks
	Blocks struct {
		MinHeight  int                `json:"min_height"`
		MaxHeight  int                `json:"max_height"`
		BlockMetas []*types.BlockMeta `json:"block_metas"`
	}

	// *********************************** Consensus ***********************************

	// Validators
	ValidatorList struct {
		BlockHeight         int                `json:"block_height"`
		BondedValidators    []*types.Validator `json:"bonded_validators"`
		UnbondingValidators []*types.Validator `json:"unbonding_validators"`
	}

	// *********************************** Events ***********************************

	// EventSubscribe
	EventSub struct {
		SubId string `json:"sub_id"`
	}

	// EventUnsubscribe
	EventUnsub struct {
		Result bool `json:"result"`
	}

	// EventPoll
	PollResponse struct {
		Events []interface{} `json:"events"`
	}

	// *********************************** Network ***********************************

	// NetworkInfo
	NetworkInfo struct {
		ClientVersion string   `json:"client_version"`
		Moniker       string   `json:"moniker"`
		Listening     bool     `json:"listening"`
		Listeners     []string `json:"listeners"`
		Peers         []*Peer  `json:"peers"`
	}

	ClientVersion struct {
		ClientVersion string `json:"client_version"`
	}

	Moniker struct {
		Moniker string `json:"moniker"`
	}

	Listening struct {
		Listening bool `json:"listening"`
	}

	Listeners struct {
		Listeners []string `json:"listeners"`
	}

	// used in Peers and BlockchainInfo
	Peer struct {
		NodeInfo   *p2p.NodeInfo `json:"node_info"`
		IsOutbound bool          `json:"is_outbound"`
	}

	// *********************************** Transactions ***********************************

	// Call or CallCode
	Call struct {
		Return  string `json:"return"`
		GasUsed int64  `json:"gas_used"`
		// TODO ...
	}
)

//------------------------------------------------------------------------------
// copied in from NameReg

type (
	NameRegEntry struct {
		Name    string `json:"name"`    // registered name for the entry
		Owner   []byte `json:"owner"`   // address that created the entry
		Data    string `json:"data"`    // data to store under this name
		Expires int    `json:"expires"` // block at which this entry expires
	}

	ResultListNames struct {
		BlockHeight int             `json:"block_height"`
		Names       []*NameRegEntry `json:"names"`
	}
)
