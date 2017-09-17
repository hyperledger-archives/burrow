// +build !arm

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

package blockchain

import (
	"time"

	"github.com/hyperledger/burrow/genesis"
)

// Pointer to the tip of the blockchain
type BlockchainTip interface {
	ChainID() string
	// All Last* references are to the block last committed
	LastBlockHeight() uint64
	LastBlockTime() time.Time
	LastBlockHash() []byte
	// Note this is the hash of the application state after the most recently committed block's transactions executed
	// and so lastBlock.Header.AppHash will be one block older than our AppHashAfterLastBlock (i.e. Tendermint closes
	// the AppHash we return from ABCI Commit into the _next_ block)
	AppHashAfterLastBlock() []byte
}

// Burrow's view on the blockchain maintained by tendermint
type Blockchain interface {
	CommitBlock(blockTime time.Time, blockHash, appHash []byte)
	GenesisDoc() genesis.GenesisDoc
	BlockchainTip
}

type blockchain struct {
	chainID               string
	genesisDoc            genesis.GenesisDoc
	lastBlockHeight       uint64
	lastBlockTime         time.Time
	lastBlockHash         []byte
	appHashAfterLastBlock []byte
}

var _ Blockchain = &blockchain{}

// Mutable pointer to blockchain tip initialised from genesis
func NewBlockchainFromGenesisDoc(genesisDoc *genesis.GenesisDoc) *blockchain {
	return &blockchain{
		chainID:               genesisDoc.ChainID,
		genesisDoc:            *genesisDoc,
		lastBlockTime:         genesisDoc.GenesisTime,
		appHashAfterLastBlock: genesisDoc.Hash(),
	}
}

func NewBlockchain(chainID string, lastBlockHeight uint64, lastBlockTime time.Time,
	lastBlockHash, appHashAfterLastBlock []byte) *blockchain {
	return &blockchain{
		chainID:               chainID,
		lastBlockHeight:       lastBlockHeight,
		lastBlockTime:         lastBlockTime,
		lastBlockHash:         lastBlockHash,
		appHashAfterLastBlock: appHashAfterLastBlock,
	}
}

func (bc *blockchain) CommitBlock(blockTime time.Time, blockHash, appHash []byte) {
	bc.lastBlockHeight += 1
	bc.lastBlockTime = blockTime
	bc.lastBlockHash = blockHash
	bc.appHashAfterLastBlock = appHash
}

func (bc *blockchain) GenesisDoc() genesis.GenesisDoc {
	return bc.genesisDoc
}

func (bc *blockchain) ChainID() string {
	return bc.chainID
}

func (bc *blockchain) LastBlockHeight() uint64 {
	return bc.lastBlockHeight
}

func (bc *blockchain) LastBlockTime() time.Time {
	return bc.lastBlockTime
}

func (bc *blockchain) LastBlockHash() []byte {
	return bc.lastBlockHash
}

func (bc *blockchain) AppHashAfterLastBlock() []byte {
	return bc.appHashAfterLastBlock
}
