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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
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

	// Returns the instance of the current validator set
	ValidatorSet() ValidatorSet
	EvaluateSortition(blockHeight uint64, prevBlockHash []byte)
	VerifySortition(prevBlockHash []byte, publicKey acm.PublicKey, info uint64, proof []byte) bool
}

// Burrow's portion of the Blockchain state
type Blockchain interface {
	// Read locker
	sync.Locker
	Root
	Tip
	// Returns an immutable copy of the tip
	Tip() Tip
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

	validatorSet ValidatorSet
	sortition    acm.Sortition
}

type blockchain struct {
	sync.RWMutex
	db dbm.DB
	*root
	*tip
}

var _ Root = &blockchain{}
var _ Tip = &blockchain{}
var _ Blockchain = &blockchain{}
var _ MutableBlockchain = &blockchain{}

func LoadOrNewBlockchain(db dbm.DB, genesisDoc *genesis.GenesisDoc,
	sortition acm.Sortition, logger *logging.Logger) (*blockchain, error) {

	logger = logger.WithScope("LoadOrNewBlockchain")
	logger.InfoMsg("Trying to load blockchain state from database",
		"database_key", stateKey)
	blockchain, err := loadBlockchain(db)
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
	} else {
		logger.InfoMsg("No existing blockchain state found in database, making new blockchain")
		blockchain = newBlockchain(db, genesisDoc)
		blockchain.validatorSet = newValidatorSet(genesisDoc.GetMaximumPower(), genesisDoc.Validators())
	}

	blockchain.sortition = sortition
	return blockchain, nil
}

// Pointer to blockchain state initialised from genesis
func newBlockchain(db dbm.DB, genesisDoc *genesis.GenesisDoc) *blockchain {

	return &blockchain{
		db:   db,
		root: NewRoot(genesisDoc),
		tip:  NewTip(genesisDoc.ChainID(), genesisDoc.GenesisTime, genesisDoc.Hash()),
	}
}

func loadBlockchain(db dbm.DB) (*blockchain, error) {
	buf := db.Get(stateKey)
	if len(buf) == 0 {
		return nil, nil
	}
	u := map[string]string{}
	err := json.Unmarshal(buf, &u)
	if err != nil {
		return nil, err
	}

	genesisDoc, err := genesis.GenesisDocFromJSON([]byte(u["genesisDoc"]))
	if err != nil {
		return nil, err
	}

	validatorSet, err := ValidatorSetFromJSON([]byte(u["validatorSet"]))
	if err != nil {
		return nil, err
	}

	lastBlockHeight, err := strconv.ParseUint(u["lastBlockHeight"], 10, 64)
	if err != nil {
		return nil, err
	}

	appHashAfterLastBlock, err := hex.DecodeString(u["appHashAfterLastBlock"])
	if err != nil {
		return nil, err
	}

	blockchain := newBlockchain(db, genesisDoc)
	blockchain.lastBlockHeight = lastBlockHeight
	blockchain.appHashAfterLastBlock = appHashAfterLastBlock
	blockchain.validatorSet = validatorSet

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
		genesisJSON, err := bc.genesisDoc.JSONBytes()
		if err != nil {
			return err
		}

		validatorSetJSON, err := bc.validatorSet.JSONBytes()
		if err != nil {
			return err
		}

		u := map[string]string{}
		u["genesisDoc"] = string(genesisJSON)
		u["validatorSet"] = string(validatorSetJSON)
		u["lastBlockHeight"] = strconv.FormatUint(bc.lastBlockHeight, 10)
		u["appHashAfterLastBlock"] = hex.EncodeToString(bc.appHashAfterLastBlock)

		encodedState, err := json.Marshal(u)
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

func (t *tip) ValidatorSet() ValidatorSet {
	return t.validatorSet
}

func (t *tip) EvaluateSortition(blockHeight uint64, prevBlockHash []byte) {

	// Check if sortition is set
	if t.sortition == nil {
		return
	}

	// check if this validator is in set or not
	if t.validatorSet.IsValidatorInSet(t.sortition.Address()) {
		return
	}

	// this validator is not in the set
	go t.sortition.Evaluate(blockHeight, prevBlockHash)
}

func (t *tip) VerifySortition(prevBlockHash []byte, publicKey acm.PublicKey, info uint64, proof []byte) bool {
	return t.sortition.Verify(prevBlockHash, publicKey, info, proof)
}
