package forensics

import (
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/types"
)

type Block struct {
	txDecoder txs.Decoder
	*types.Block
}

func NewBlock(txDecoder txs.Decoder, block *types.Block) *Block {
	return &Block{
		txDecoder: txDecoder,
		Block:     block,
	}
}

func (b *Block) Transactions(iter func(txs.Tx) (stop bool)) (stopped bool, err error) {
	for i := 0; i < len(b.Txs); i++ {
		tx, err := b.txDecoder.DecodeTx(b.Txs[i])
		if err != nil {
			return false, err
		}
		if iter(tx) {
			return true, nil
		}
	}
	return false, nil
}
