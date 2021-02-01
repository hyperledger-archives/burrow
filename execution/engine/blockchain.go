package engine

import (
	"encoding/binary"
	"time"

	"github.com/hyperledger/burrow/execution/errors"
)

type TestBlockchain struct {
	BlockHeight uint64
	BlockTime   time.Time
}

func (b *TestBlockchain) LastBlockHeight() uint64 {
	return b.BlockHeight
}

func (b *TestBlockchain) LastBlockTime() time.Time {
	return b.BlockTime
}

func (b *TestBlockchain) BlockHash(height uint64) ([]byte, error) {
	if height > b.BlockHeight {
		return nil, errors.Codes.InvalidBlockNumber
	}
	bs := make([]byte, 32)
	binary.BigEndian.PutUint64(bs[24:], height)
	return bs, nil
}
