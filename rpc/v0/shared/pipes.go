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

package shared

// Shared extension methods for Pipe and its derivatives

import (
	core_types "github.com/monax/eris-db/core/types"
	"github.com/monax/eris-db/definitions"
	tendermint_types "github.com/tendermint/tendermint/types"
)

func BlockchainInfo(pipe definitions.Pipe) *core_types.BlockchainInfo {
	latestHeight := pipe.Blockchain().Height()

	var latestBlockMeta *tendermint_types.BlockMeta

	if latestHeight != 0 {
		latestBlockMeta = pipe.Blockchain().BlockMeta(latestHeight)
	}

	return &core_types.BlockchainInfo{
		ChainId:           pipe.Blockchain().ChainId(),
		GenesisHash:       pipe.GenesisHash(),
		LatestBlockHeight: latestHeight,
		LatestBlock:       latestBlockMeta,
	}
}
