// Copyright 2019 Monax Industries Limited
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

package bcm

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

func (b *Block) Transactions(iter func(*txs.Envelope) (stop bool)) (stopped bool, err error) {
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
