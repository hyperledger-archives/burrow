package bcm

import (
	"fmt"
	"runtime/debug"

	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

type BlockStore struct {
	txDecoder txs.Decoder
	state.BlockStore
}

func NewBlockStore(blockStore state.BlockStore) *BlockStore {
	return &BlockStore{
		txDecoder:  txs.NewProtobufCodec(),
		BlockStore: blockStore,
	}
}

func NewBlockExplorer(dbBackendType dbm.BackendType, dbDir string) (*BlockStore, error) {
	db, err := dbm.NewDB("blockstore", dbBackendType, dbDir)
	if err != nil {
		return nil, fmt.Errorf("could not create BlockExplorer: %w", err)
	}
	return NewBlockStore(store.NewBlockStore(db)), nil
}

func (bs *BlockStore) Block(height int64) (_ *Block, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("BlockStore.Block(): could not get block at height %v: %v", height, r)
		}
	}()

	tmBlock := bs.LoadBlock(height)
	if tmBlock == nil {
		return nil, fmt.Errorf("could not pull block at height: %v", height)
	}
	return NewBlock(bs.txDecoder, tmBlock), nil
}

func (bs *BlockStore) BlockMeta(height int64) (_ *types.BlockMeta, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("BlockStore.BlockMeta() could not get BlockMeta at height %d: %v\n%s",
				height, r, debug.Stack())
		}
	}()
	return bs.LoadBlockMeta(height), nil
}

// Iterate over blocks between start (inclusive) and end (exclusive)
func (bs *BlockStore) Blocks(start, end int64, iter func(*Block) error) error {
	if end > 0 && start >= end {
		return fmt.Errorf("end height must be strictly greater than start height")
	}
	if start <= 0 {
		// From first block
		start = 1
	}
	if end < 0 {
		// -1 means include the very last block so + 1 for offset
		end = bs.Height() + end + 1
	}

	for height := start; height <= end; height++ {
		block, err := bs.Block(height)
		if err != nil {
			return err
		}
		err = iter(block)
		if err != nil {
			return err
		}
	}

	return nil
}
