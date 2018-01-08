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

import (
	tm_types "github.com/tendermint/tendermint/types"
)

type (
	// *********************************** Blockchain ***********************************

	// BlockchainInfo
	BlockchainInfo struct {
		ChainId           string              `json:"chain_id"`
		GenesisHash       []byte              `json:"genesis_hash"`
		LatestBlockHeight uint64              `json:"latest_block_height"`
		LatestBlock       *tm_types.BlockMeta `json:"latest_block"`
	}

	// Genesis hash
	GenesisHash struct {
		Hash []byte `json:"hash"`
	}

	// Get the latest
	LatestBlockHeight struct {
		Height uint64 `json:"height"`
	}

	ChainId struct {
		ChainId string `json:"chain_id"`
	}

	// GetBlocks
	Blocks struct {
		MinHeight  uint64                `json:"min_height"`
		MaxHeight  uint64                `json:"max_height"`
		BlockMetas []*tm_types.BlockMeta `json:"block_metas"`
	}

	// *********************************** Consensus ***********************************

	// Validators
	ValidatorList struct {
		BlockHeight         uint64                `json:"block_height"`
		BondedValidators    []*tm_types.Validator `json:"bonded_validators"`
		UnbondingValidators []*tm_types.Validator `json:"unbonding_validators"`
	}
)
