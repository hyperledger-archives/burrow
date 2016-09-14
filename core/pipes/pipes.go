package pipes

// Shared extension methods for Pipe and its derivatives

import (
	core_types "github.com/eris-ltd/eris-db/core/types"
	"github.com/eris-ltd/eris-db/definitions"
	tendermint_types "github.com/tendermint/tendermint/types"
)

func BlockchainInfo(pipe definitions.Pipe) *core_types.BlockchainInfo {
	latestHeight := pipe.Blockchain().Height()

	var latestBlockMeta *tendermint_types.BlockMeta

	if latestHeight != 0 {
		latestBlockMeta = pipe.Blockchain().BlockMeta(latestHeight)
	}

	return &core_types.BlockchainInfo{
		ChainId: pipe.Blockchain().ChainId(),
		GenesisHash: pipe.GenesisHash(),
		LatestBlockHeight: latestHeight,
		LatestBlock: latestBlockMeta,
	}
}
