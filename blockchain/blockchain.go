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
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	dbm "github.com/tendermint/tmlibs/db"
)

var stateKey = []byte("BlockchainState")

// Immutable Root of blockchain
type Root interface {
	// GenesisHash precomputed from GenesisDoc
	GenesisHash() []byte
	GenesisDoc() genesis.GenesisDoc
}

// Immutable pointer to the current tip of the blockchain
type Tip interface {
	// ChainID precomputed from GenesisDoc
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

// Burrow's portion of the Blockchain state
type Blockchain interface {
	// Read locker
	sync.Locker
	Root
	Tip
	// Returns an immutable copy of the tip
	Tip() Tip
	// Returns a copy of the current validator set
	Validators() []acm.Validator
}

type MutableBlockchain interface {
	Blockchain
	CommitBlock(blockTime time.Time, blockHash, appHash []byte) error
}

type root struct {
	genesisHash []byte
	genesisDoc  genesis.GenesisDoc
}

type tip struct {
	chainID               string
	lastBlockHeight       uint64
	lastBlockTime         time.Time
	lastBlockHash         []byte
	appHashAfterLastBlock []byte
}

type blockchain struct {
	sync.RWMutex
	db dbm.DB
	*root
	*tip
	validators []acm.Validator
}

var _ Root = &blockchain{}
var _ Tip = &blockchain{}
var _ Blockchain = &blockchain{}
var _ MutableBlockchain = &blockchain{}

type PersistedState struct {
	AppHashAfterLastBlock []byte
	LastBlockHeight       uint64
	GenesisDoc            genesis.GenesisDoc
}

func LoadOrNewBlockchain(db dbm.DB, genesisDoc *genesis.GenesisDoc,
	logger *logging.Logger) (*blockchain, error) {

	logger = logger.WithScope("LoadOrNewBlockchain")
	logger.InfoMsg("Trying to load blockchain state from database",
		"database_key", stateKey)
	blockchain, err := LoadBlockchain(db)
	if err != nil {
		return nil, fmt.Errorf("error loading blockchain state from database: %v", err)
	}
	if blockchain != nil {
		dbHash := blockchain.genesisDoc.Hash()
		argHash := genesisDoc.Hash()
		if !bytes.Equal(dbHash, argHash) {
			return nil, fmt.Errorf("GenesisDoc passed to LoadOrNewBlockchain has hash: 0x%X, which does not "+
				"match the one found in database: 0x%X", argHash, dbHash)
		}
		return blockchain, nil
	}

	logger.InfoMsg("No existing blockchain state found in database, making new blockchain")
	return NewBlockchain(db, genesisDoc), nil
}

// Pointer to blockchain state initialised from genesis
func NewBlockchain(db dbm.DB, genesisDoc *genesis.GenesisDoc) *blockchain {
	var validators []acm.Validator
	for _, gv := range genesisDoc.Validators {
		validators = append(validators, acm.ConcreteValidator{
			PublicKey: gv.PublicKey,
			Power:     uint64(gv.Amount),
		}.Validator())
	}
	root := NewRoot(genesisDoc)
	return &blockchain{
		db:         db,
		root:       root,
		tip:        NewTip(genesisDoc.ChainID(), root.genesisDoc.GenesisTime, root.genesisHash),
		validators: validators,
	}
}

func LoadBlockchain(db dbm.DB) (*blockchain, error) {
	buf := db.Get(stateKey)
	if len(buf) == 0 {
		return nil, nil
	}
	persistedState, err := Decode(buf)
	if err != nil {
		return nil, err
	}
	blockchain := NewBlockchain(db, &persistedState.GenesisDoc)
	blockchain.lastBlockHeight = persistedState.LastBlockHeight
	blockchain.appHashAfterLastBlock = persistedState.AppHashAfterLastBlock
	return blockchain, nil
}

func NewRoot(genesisDoc *genesis.GenesisDoc) *root {
	return &root{
		genesisHash: genesisDoc.Hash(),
		genesisDoc:  *genesisDoc,
	}
}

// Create genesis Tip
func NewTip(chainID string, genesisTime time.Time, genesisHash []byte) *tip {
	return &tip{
		chainID:               chainID,
		lastBlockTime:         genesisTime,
		appHashAfterLastBlock: genesisHash,
	}
}

func (bc *blockchain) CommitBlock(blockTime time.Time, blockHash, appHash []byte) error {
	bc.Lock()
	defer bc.Unlock()
	bc.lastBlockHeight += 1
	bc.lastBlockTime = blockTime
	bc.lastBlockHash = blockHash
	bc.appHashAfterLastBlock = appHash
	return bc.save()
}

func (bc *blockchain) save() error {
	if bc.db != nil {
		encodedState, err := bc.Encode()
		if err != nil {
			return err
		}
		bc.db.SetSync(stateKey, encodedState)
	}
	return nil
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

func (bc *blockchain) Validators() []acm.Validator {
	bc.RLock()
	defer bc.RUnlock()
	vs := make([]acm.Validator, len(bc.validators))
	for i, v := range bc.validators {
		vs[i] = v
	}
	return vs
}

func (bc *blockchain) Encode() ([]byte, error) {
	persistedState := &PersistedState{
		GenesisDoc:            bc.genesisDoc,
		AppHashAfterLastBlock: bc.appHashAfterLastBlock,
		LastBlockHeight:       bc.lastBlockHeight,
	}
	encodedState, err := json.Marshal(persistedState)
	if err != nil {
		return nil, err
	}
	return encodedState, nil
}

func Decode(encodedState []byte) (*PersistedState, error) {
	persistedState := new(PersistedState)
	err := json.Unmarshal(encodedState, persistedState)
	if err != nil {
		return nil, err
	}
	return persistedState, nil
}

func (r *root) GenesisHash() []byte {
	return r.genesisHash
}

func (r *root) GenesisDoc() genesis.GenesisDoc {
	return r.genesisDoc
}

func (t *tip) ChainID() string {
	return t.chainID
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
