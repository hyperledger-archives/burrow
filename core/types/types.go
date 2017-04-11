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

package types

// TODO: [ben] this is poorly constructed but copied over
// from burrow/burrow/pipe/types to make incremental changes and allow
// for a discussion around the proper defintion of the needed types.

import (
	// NodeInfo (drop this!)
	"github.com/tendermint/tendermint/types"

	account "github.com/hyperledger/burrow/account"
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
