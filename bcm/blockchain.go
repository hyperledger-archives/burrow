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

package bcm

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/tendermint/tendermint/types"

	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	amino "github.com/tendermint/go-amino"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var stateKey = []byte("BlockchainState")

type BlockchainInfo interface {
	GenesisHash() []byte
	GenesisDoc() genesis.GenesisDoc
	ChainID() string
	LastBlockHeight() uint64
	LastBlockTime() time.Time
	LastCommitTime() time.Time
	LastCommitDuration() time.Duration
	LastBlockHash() []byte
	AppHashAfterLastBlock() []byte
	// Gets the BlockHash at a height (or nil if no BlockStore mounted or block could not be found)
	BlockHash(height uint64) []byte
	// GetBlockHash returns	hash of the specific block
	GetBlockHeader(blockNumber uint64) (*types.Header, error)
}

type Blockchain struct {
	sync.RWMutex
	db                    dbm.DB
	genesisHash           []byte
	genesisDoc            genesis.GenesisDoc
	chainID               string
	lastBlockHeight       uint64
	lastBlockTime         time.Time
	lastBlockHash         []byte
	lastCommitTime        time.Time
	lastCommitDuration    time.Duration
	appHashAfterLastBlock []byte
	blockStore            *BlockStore
}

var _ BlockchainInfo = &Blockchain{}

type PersistedState struct {
	AppHashAfterLastBlock []byte
	LastBlockHeight       uint64
	GenesisHash           []byte
}

// LoadOrNewBlockchain returns true if state already exists
func LoadOrNewBlockchain(db dbm.DB, genesisDoc *genesis.GenesisDoc, logger *logging.Logger) (bool, *Blockchain, error) {
	logger = logger.WithScope("LoadOrNewBlockchain")
	logger.InfoMsg("Trying to load blockchain state from database",
		"database_key", stateKey)
	bc, err := loadBlockchain(db, genesisDoc)
	if err != nil {
		return false, nil, fmt.Errorf("error loading blockchain state from database: %v", err)
	}
	if bc != nil {
		dbHash := bc.genesisHash
		argHash := genesisDoc.Hash()
		if !bytes.Equal(dbHash, argHash) {
			return false, nil, fmt.Errorf("GenesisDoc passed to LoadOrNewBlockchain has hash: 0x%X, which does not "+
				"match the one found in database: 0x%X, database genesis:\n%v\npassed genesis:\n%v\n",
				argHash, dbHash, bc.genesisDoc.JSONString(), genesisDoc.JSONString())
		}
		return true, bc, nil
	}

	logger.InfoMsg("No existing blockchain state found in database, making new blockchain")
	return false, NewBlockchain(db, genesisDoc), nil
}

// NewBlockchain returns a pointer to blockchain state initialised from genesis
func NewBlockchain(db dbm.DB, genesisDoc *genesis.GenesisDoc) *Blockchain {
	bc := &Blockchain{
		db:                    db,
		genesisHash:           genesisDoc.Hash(),
		genesisDoc:            *genesisDoc,
		chainID:               genesisDoc.ChainID(),
		lastBlockTime:         genesisDoc.GenesisTime,
		appHashAfterLastBlock: genesisDoc.Hash(),
	}
	return bc
}

func GetSyncInfo(blockchain BlockchainInfo) *SyncInfo {
	return &SyncInfo{
		LatestBlockHeight:   blockchain.LastBlockHeight(),
		LatestBlockHash:     blockchain.LastBlockHash(),
		LatestAppHash:       blockchain.AppHashAfterLastBlock(),
		LatestBlockTime:     blockchain.LastBlockTime(),
		LatestBlockSeenTime: blockchain.LastCommitTime(),
		LatestBlockDuration: blockchain.LastCommitDuration(),
	}
}

func loadBlockchain(db dbm.DB, genesisDoc *genesis.GenesisDoc) (*Blockchain, error) {
	buf := db.Get(stateKey)
	if len(buf) == 0 {
		return nil, nil
	}
	bc, err := decodeBlockchain(buf, genesisDoc)
	if err != nil {
		return nil, err
	}
	bc.db = db
	return bc, nil
}

func (bc *Blockchain) CommitBlock(blockTime time.Time, blockHash, appHash []byte) error {
	return bc.CommitBlockAtHeight(blockTime, blockHash, appHash, bc.lastBlockHeight+1)
}

func (bc *Blockchain) CommitBlockAtHeight(blockTime time.Time, blockHash, appHash []byte, height uint64) error {
	bc.Lock()
	defer bc.Unlock()
	// Checkpoint on the _previous_ block. If we die, this is where we will resume since we know all intervening state
	// has been written successfully since we are committing the next block.
	// If we fall over we can resume a safe committed state and Tendermint will catch us up
	err := bc.save()
	if err != nil {
		return err
	}
	bc.lastCommitDuration = blockTime.Sub(bc.lastBlockTime)
	bc.lastBlockHeight = height
	bc.lastBlockTime = blockTime
	bc.lastBlockHash = blockHash
	bc.appHashAfterLastBlock = appHash
	bc.lastCommitTime = time.Now().UTC()
	return nil
}

func (bc *Blockchain) CommitWithAppHash(appHash []byte) error {
	bc.appHashAfterLastBlock = appHash
	bc.Lock()
	defer bc.Unlock()

	return bc.save()
}

func (bc *Blockchain) save() error {
	if bc.db != nil {
		encodedState, err := bc.Encode()
		if err != nil {
			return err
		}
		bc.db.SetSync(stateKey, encodedState)
	}
	return nil
}

var cdc = amino.NewCodec()

func (bc *Blockchain) Encode() ([]byte, error) {
	persistedState := &PersistedState{
		GenesisHash:           bc.genesisDoc.Hash(),
		AppHashAfterLastBlock: bc.appHashAfterLastBlock,
		LastBlockHeight:       bc.lastBlockHeight,
	}
	encodedState, err := cdc.MarshalBinaryBare(persistedState)
	if err != nil {
		return nil, err
	}
	return encodedState, nil
}

func decodeBlockchain(encodedState []byte, genesisDoc *genesis.GenesisDoc) (*Blockchain, error) {
	persistedState := new(PersistedState)
	err := cdc.UnmarshalBinaryBare(encodedState, persistedState)
	if err != nil {
		return nil, err
	}

	bc := NewBlockchain(nil, genesisDoc)
	bc.genesisHash = persistedState.GenesisHash
	//bc.lastBlockHeight = persistedState.LastBlockHeight
	bc.lastBlockHeight = persistedState.LastBlockHeight
	bc.appHashAfterLastBlock = persistedState.AppHashAfterLastBlock
	return bc, nil
}

func (bc *Blockchain) GenesisHash() []byte {
	return bc.genesisHash
}

func (bc *Blockchain) GenesisDoc() genesis.GenesisDoc {
	return bc.genesisDoc
}

func (bc *Blockchain) ChainID() string {
	return bc.chainID
}

func (bc *Blockchain) LastBlockHeight() uint64 {
	if bc == nil {
		return 0
	}
	bc.RLock()
	defer bc.RUnlock()
	return bc.lastBlockHeight
}

func (bc *Blockchain) LastBlockTime() time.Time {
	bc.RLock()
	defer bc.RUnlock()
	return bc.lastBlockTime
}

func (bc *Blockchain) LastCommitTime() time.Time {
	bc.RLock()
	defer bc.RUnlock()
	return bc.lastCommitTime
}

func (bc *Blockchain) LastCommitDuration() time.Duration {
	bc.RLock()
	defer bc.RUnlock()
	return bc.lastCommitDuration
}

func (bc *Blockchain) LastBlockHash() []byte {
	bc.RLock()
	defer bc.RUnlock()
	return bc.lastBlockHash
}

func (bc *Blockchain) AppHashAfterLastBlock() []byte {
	bc.RLock()
	defer bc.RUnlock()
	return bc.appHashAfterLastBlock
}

// Tendermint block access

func (bc *Blockchain) SetBlockStore(bs *BlockStore) {
	bc.blockStore = bs
}

func (bc *Blockchain) BlockHash(height uint64) []byte {
	header, err := bc.GetBlockHeader(height)
	if err != nil {
		return nil
	}
	return header.Hash()
}

func (bc *Blockchain) GetBlockHeader(height uint64) (*types.Header, error) {
	const errHeader = "GetBlockHeader():"
	if bc == nil {
		return nil, fmt.Errorf("%s could not get block hash because Blockchain has not been given access to "+
			"tendermint BlockStore", errHeader)
	}
	blockMeta, err := bc.blockStore.BlockMeta(int64(height))
	if err != nil {
		return nil, fmt.Errorf("%s could not get BlockMeta: %v", errHeader, err)
	}
	return &blockMeta.Header, nil
}
