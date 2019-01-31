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
	"math/big"
	"time"

	"sync"

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
	LastBlockHash() []byte
	AppHashAfterLastBlock() []byte
	// GetBlockHash returns	hash of the specific block
	GetBlockHash(blockNumber uint64) ([]byte, error)
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
	appHashAfterLastBlock []byte
	blockStore            *BlockStore
}

var _ BlockchainInfo = &Blockchain{}

type PersistedState struct {
	AppHashAfterLastBlock []byte
	LastBlockHeight       uint64
	GenesisDoc            genesis.GenesisDoc
}

func LoadOrNewBlockchain(db dbm.DB, genesisDoc *genesis.GenesisDoc, logger *logging.Logger) (*Blockchain, error) {
	logger = logger.WithScope("LoadOrNewBlockchain")
	logger.InfoMsg("Trying to load blockchain state from database",
		"database_key", stateKey)
	bc, err := loadBlockchain(db)
	if err != nil {
		return nil, fmt.Errorf("error loading blockchain state from database: %v", err)
	}
	if bc != nil {
		dbHash := bc.genesisDoc.Hash()
		argHash := genesisDoc.Hash()
		if !bytes.Equal(dbHash, argHash) {
			return nil, fmt.Errorf("GenesisDoc passed to LoadOrNewBlockchain has hash: 0x%X, which does not "+
				"match the one found in database: 0x%X, database genesis:\n%v\npassed genesis:\n%v\n",
				argHash, dbHash, bc.genesisDoc.JSONString(), genesisDoc.JSONString())
		}
		return bc, nil
	}

	logger.InfoMsg("No existing blockchain state found in database, making new blockchain")
	return NewBlockchain(db, genesisDoc), nil
}

// Pointer to blockchain state initialised from genesis
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

func loadBlockchain(db dbm.DB) (*Blockchain, error) {
	buf := db.Get(stateKey)
	if len(buf) == 0 {
		return nil, nil
	}
	bc, err := DecodeBlockchain(buf)
	if err != nil {
		return nil, err
	}
	bc.db = db
	return bc, nil
}

func (bc *Blockchain) CommitBlock(blockTime time.Time, blockHash,
	appHash []byte) (totalPowerChange, totalFlow *big.Int, err error) {
	return bc.CommitBlockAtHeight(blockTime, blockHash, appHash, bc.lastBlockHeight+1)
}

func (bc *Blockchain) CommitBlockAtHeight(blockTime time.Time, blockHash, appHash []byte,
	height uint64) (totalPowerChange, totalFlow *big.Int, err error) {
	bc.Lock()
	defer bc.Unlock()
	// Checkpoint on the _previous_ block. If we die, this is where we will resume since we know all intervening state
	// has been written successfully since we are committing the next block.
	// If we fall over we can resume a safe committed state and Tendermint will catch us up
	err = bc.save()
	if err != nil {
		return
	}
	bc.lastBlockHeight = height
	bc.lastBlockTime = blockTime
	bc.lastBlockHash = blockHash
	bc.appHashAfterLastBlock = appHash
	bc.lastCommitTime = time.Now().UTC()
	return
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
		GenesisDoc:            bc.genesisDoc,
		AppHashAfterLastBlock: bc.appHashAfterLastBlock,
		LastBlockHeight:       bc.lastBlockHeight,
	}
	encodedState, err := cdc.MarshalBinaryBare(persistedState)
	if err != nil {
		return nil, err
	}
	return encodedState, nil
}

func DecodeBlockchain(encodedState []byte) (*Blockchain, error) {
	persistedState := new(PersistedState)
	err := cdc.UnmarshalBinaryBare(encodedState, persistedState)
	if err != nil {
		return nil, err
	}
	bc := NewBlockchain(nil, &persistedState.GenesisDoc)
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

func (bc *Blockchain) GetBlockHash(height uint64) ([]byte, error) {
	const errHeader = "GetBlockHash():"
	if bc == nil {
		return nil, fmt.Errorf("%s could not get block hash because Blockchain has not been given access to "+
			"tendermint BlockStore", errHeader)
	}
	blockMeta, err := bc.blockStore.BlockMeta(int64(height))
	if err != nil {
		return nil, fmt.Errorf("%s could not get BlockMeta: %v", errHeader, err)
	}
	return blockMeta.Header.Hash(), nil
}
