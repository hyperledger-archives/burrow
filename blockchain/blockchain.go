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

	"sync"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/genesis"
)

// Immutable Root of blockchain
type Root interface {
	// ChainID precomputed from GenesisDoc
	ChainID() string
	// GenesisHash precomputed from GenesisDoc
	GenesisHash() []byte
	GenesisDoc() genesis.GenesisDoc
}

// Immutable pointer to the current tip of the blockchain
type Tip interface {
	// All Last* references are to the block last committed
	LastBlockHeight() uint64
	LastBlockTime() time.Time
	LastBlockHash() []byte
	// Note this is the hash of the application state after the most recently committed block's transactions executed
	// and so lastBlock.Header.AppHash will be one block older than our AppHashAfterLastBlock (i.e. Tendermint closes
	// the AppHash we return from ABCI Commit into the _next_ block)
	AppHashAfterLastBlock() []byte
}

// Burrow's portion of the Blockchain state
type Blockchain interface {
	Root
	Tip
	// Returns an immutable copy of the tip
	Tip() Tip
	// Returns a copy of the current validator set
	Validators() []*acm.Validator
}

type MutableBlockchain interface {
	Blockchain
	CommitBlock(blockTime time.Time, blockHash, appHash []byte)
}

type root struct {
	chainID     string
	genesisHash []byte
	genesisDoc  genesis.GenesisDoc
}

type tip struct {
	lastBlockHeight       uint64
	lastBlockTime         time.Time
	lastBlockHash         []byte
	appHashAfterLastBlock []byte
}

type blockchain struct {
	sync.RWMutex
	*root
	*tip
	validators []acm.Validator
}

var _ Root = &blockchain{}
var _ Tip = &blockchain{}
var _ Blockchain = &blockchain{}
var _ MutableBlockchain = &blockchain{}

// Pointer to blockchain state initialised from genesis
func NewBlockchain(genesisDoc *genesis.GenesisDoc) *blockchain {
	var validators []acm.Validator
	for _, gv := range genesisDoc.Validators {
		validators = append(validators, acm.ConcreteValidator{
			PubKey: gv.PubKey,
			Power:  uint64(gv.Amount),
		}.Validator())
	}
	root := NewRoot(genesisDoc)
	return &blockchain{
		root: root,
		tip: &tip{
			lastBlockTime:         root.genesisDoc.GenesisTime,
			appHashAfterLastBlock: root.genesisHash,
		},
		validators: validators,
	}
}

func NewRoot(genesisDoc *genesis.GenesisDoc) *root {
	return &root{
		chainID:     genesisDoc.ChainID(),
		genesisHash: genesisDoc.Hash(),
		genesisDoc:  *genesisDoc,
	}
}

// Create
func NewTip(lastBlockHeight uint64, lastBlockTime time.Time, lastBlockHash []byte, appHashAfterLastBlock []byte) *tip {
	return &tip{
		lastBlockHeight:       lastBlockHeight,
		lastBlockTime:         lastBlockTime,
		lastBlockHash:         lastBlockHash,
		appHashAfterLastBlock: appHashAfterLastBlock,
	}
}

func (bc *blockchain) CommitBlock(blockTime time.Time, blockHash, appHash []byte) {
	bc.Lock()
	defer bc.Unlock()
	bc.lastBlockHeight += 1
	bc.lastBlockTime = blockTime
	bc.lastBlockHash = blockHash
	bc.appHashAfterLastBlock = appHash
}

func (bc *blockchain) Root() Root {
	return bc.root
}

func (bc *blockchain) Tip() Tip {
	bc.RLock()
	defer bc.RUnlock()
	t := *bc.tip
	return &t
}

func (bc *blockchain) Validators() []*acm.Validator {
	bc.RLock()
	defer bc.RUnlock()
	vs := make([]*acm.Validator, len(bc.validators))
	for i, v := range bc.validators {
		vs[i] = &v
	}
	return vs
}

func (r *root) ChainID() string {
	return r.chainID
}

func (r *root) GenesisHash() []byte {
	return r.genesisHash
}

func (r *root) GenesisDoc() genesis.GenesisDoc {
	return r.genesisDoc
}

func (t *tip) LastBlockHeight() uint64 {
	return t.lastBlockHeight
}

func (t *tip) LastBlockTime() time.Time {
	return t.lastBlockTime
}

func (t *tip) LastBlockHash() []byte {
	return t.lastBlockHash
}

func (t *tip) AppHashAfterLastBlock() []byte {
	return t.appHashAfterLastBlock
}
