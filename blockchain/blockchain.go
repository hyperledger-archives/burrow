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
	"time"

	"sync"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	dbm "github.com/tendermint/tmlibs/db"
)

// Blocks to average validator power over
const DefaultValidatorsWindowSize = 10

var stateKey = []byte("BlockchainState")

type BlockchainInfo interface {
	GenesisHash() []byte
	GenesisDoc() genesis.GenesisDoc
	ChainID() string
	LastBlockHeight() uint64
	LastBlockTime() time.Time
	LastBlockHash() []byte
	AppHashAfterLastBlock() []byte
}

type Root struct {
	genesisHash []byte
	genesisDoc  genesis.GenesisDoc
}

type Tip struct {
	chainID               string
	lastBlockHeight       uint64
	lastBlockTime         time.Time
	lastBlockHash         []byte
	appHashAfterLastBlock []byte
	validators            Validators
	validatorsWindow      ValidatorsWindow
}

type Blockchain struct {
	*Root
	*Tip
	sync.RWMutex
	db dbm.DB
}

type PersistedState struct {
	AppHashAfterLastBlock []byte
	LastBlockHeight       uint64
	GenesisDoc            genesis.GenesisDoc
}

func LoadOrNewBlockchain(db dbm.DB, genesisDoc *genesis.GenesisDoc,
	logger *logging.Logger) (*Blockchain, error) {

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
				"match the one found in database: 0x%X", argHash, dbHash)
		}
		return bc, nil
	}

	logger.InfoMsg("No existing blockchain state found in database, making new blockchain")
	return newBlockchain(db, genesisDoc), nil
}

// Pointer to blockchain state initialised from genesis
func newBlockchain(db dbm.DB, genesisDoc *genesis.GenesisDoc) *Blockchain {
	bc := &Blockchain{
		db:   db,
		Root: NewRoot(genesisDoc),
		Tip:  NewTip(genesisDoc.ChainID(), NewRoot(genesisDoc).genesisDoc.GenesisTime, NewRoot(genesisDoc).genesisHash),
	}
	for _, gv := range genesisDoc.Validators {
		bc.validators.AlterPower(gv.PublicKey, gv.Amount)
	}
	return bc
}

func loadBlockchain(db dbm.DB) (*Blockchain, error) {
	buf := db.Get(stateKey)
	if len(buf) == 0 {
		return nil, nil
	}
	persistedState, err := Decode(buf)
	if err != nil {
		return nil, err
	}
	bc := newBlockchain(db, &persistedState.GenesisDoc)
	bc.lastBlockHeight = persistedState.LastBlockHeight
	bc.appHashAfterLastBlock = persistedState.AppHashAfterLastBlock
	return bc, nil
}

func NewRoot(genesisDoc *genesis.GenesisDoc) *Root {
	return &Root{
		genesisHash: genesisDoc.Hash(),
		genesisDoc:  *genesisDoc,
	}
}

// Create genesis Tip
func NewTip(chainID string, genesisTime time.Time, genesisHash []byte) *Tip {
	return &Tip{
		chainID:               chainID,
		lastBlockTime:         genesisTime,
		appHashAfterLastBlock: genesisHash,
		validators:            NewValidators(),
		validatorsWindow:      NewValidatorsWindow(DefaultValidatorsWindowSize),
	}
}

func (bc *Blockchain) CommitBlock(blockTime time.Time, blockHash, appHash []byte) error {
	bc.Lock()
	defer bc.Unlock()
	bc.lastBlockHeight += 1
	bc.lastBlockTime = blockTime
	bc.lastBlockHash = blockHash
	bc.appHashAfterLastBlock = appHash
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

func (bc *Blockchain) Encode() ([]byte, error) {
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

func (r *Root) GenesisHash() []byte {
	return r.genesisHash
}

func (r *Root) GenesisDoc() genesis.GenesisDoc {
	return r.genesisDoc
}

func (t *Tip) ChainID() string {
	return t.chainID
}

func (t *Tip) LastBlockHeight() uint64 {
	return t.lastBlockHeight
}

func (t *Tip) LastBlockTime() time.Time {
	return t.lastBlockTime
}

func (t *Tip) LastBlockHash() []byte {
	return t.lastBlockHash
}

func (t *Tip) AppHashAfterLastBlock() []byte {
	return t.appHashAfterLastBlock
}

func (t *Tip) IterateValidators(iter func(publicKey crypto.PublicKey, power uint64) (stop bool)) (stopped bool) {
	return t.validators.Iterate(iter)
}

func (t *Tip) NumValidators() int {
	return t.validators.Length()
}
