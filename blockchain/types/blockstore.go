package types

import tendermint_types "github.com/tendermint/tendermint/types"

type BlockStore interface {
	Height() int
	BlockMeta(height int) *tendermint_types.BlockMeta
	Block(height int) *tendermint_types.Block
}
