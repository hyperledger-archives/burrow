package forensics

import (
	"fmt"

	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/blockchain"
	"github.com/tendermint/tendermint/libs/db"
)

type BlockExplorer struct {
	txDecoder txs.Decoder
	*blockchain.BlockStore
}

func NewBlockExplorer(dbBackendType db.DBBackendType, dbDir string) *BlockExplorer {
	return &BlockExplorer{
		txDecoder:  txs.NewAminoCodec(),
		BlockStore: blockchain.NewBlockStore(tendermint.DBProvider("blockstore", dbBackendType, dbDir)),
	}
}

func (be *BlockExplorer) Block(height int64) (block *Block, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("could not get block at height %v: %v", height, r)
		}
	}()

	tmBlock := be.LoadBlock(height)
	if tmBlock == nil {
		return nil, fmt.Errorf("could not pull block at height: %v", height)
	}
	return NewBlock(be.txDecoder, tmBlock), nil
}

// Iterate over blocks between start (inclusive) and end (exclusive)
func (be *BlockExplorer) Blocks(start, end int64, iter func(*Block) (stop bool)) (stopped bool, err error) {
	if end > 0 && start >= end {
		return false, fmt.Errorf("end height must be strictly greater than start height")
	}
	if start <= 0 {
		// From first block
		start = 1
	}
	if end < 0 {
		// -1 means include the very last block so + 1 for offset
		end = be.Height() + end + 1
	}

	for height := start; height <= end; height++ {
		block, err := be.Block(height)
		if err != nil {
			return false, err
		}
		if iter(block) {
			return true, nil
		}
	}

	return false, nil
}
