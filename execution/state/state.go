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

package state

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/proposal"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/storage"
	"github.com/hyperledger/burrow/txs"
	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	DefaultValidatorsWindowSize = 10
	defaultCacheCapacity        = 1024
	uint64Length                = 8
	// Prefix under which the versioned merkle state tree resides - tracking previous versions of history
	forestPrefix = "f"
)

// Implements account and blockchain state
var _ acmstate.IterableReader = &State{}
var _ names.IterableReader = &State{}
var _ Updatable = &writeState{}

type KeyFormatStore struct {
	Account   *storage.MustKeyFormat
	Storage   *storage.MustKeyFormat
	Name      *storage.MustKeyFormat
	Proposal  *storage.MustKeyFormat
	Validator *storage.MustKeyFormat
	Event     *storage.MustKeyFormat
	TxHash    *storage.MustKeyFormat
}

var keys = KeyFormatStore{
	// AccountAddress -> Account
	Account: storage.NewMustKeyFormat("a", crypto.AddressLength),
	// AccountAddress, Key -> Value
	Storage: storage.NewMustKeyFormat("s", crypto.AddressLength, binary.Word256Length),
	// Name -> Entry
	Name: storage.NewMustKeyFormat("n", storage.VariadicSegmentLength),
	// ProposalHash -> Proposal
	Proposal: storage.NewMustKeyFormat("p", sha256.Size),
	// ValidatorAddress -> Power
	Validator: storage.NewMustKeyFormat("v", crypto.AddressLength),
	// Height, EventIndex -> StreamEvent
	Event: storage.NewMustKeyFormat("t", uint64Length, uint64Length),
	// TxHash -> TxHeight, TxIndex
	TxHash: storage.NewMustKeyFormat("th", txs.HashLength),
}

func init() {
	err := storage.EnsureKeyFormatStore(keys)
	if err != nil {
		panic(fmt.Errorf("KeyFormatStore is invalid: %v", err))
	}
}

type Updatable interface {
	acmstate.Writer
	names.Writer
	proposal.Writer
	validator.Writer
	AddBlock(blockExecution *exec.BlockExecution) error
}

// Wraps state to give access to writer methods
type writeState struct {
	forest       *storage.MutableForest
	accountStats acmstate.AccountStats
	ring         *validator.Ring
}

type ReadState struct {
	Forest storage.ForestReader
	validator.History
}

// Writers to state are responsible for calling State.Lock() before calling
type State struct {
	sync.Mutex
	db      dbm.DB
	cacheDB *storage.CacheDB
	ReadState
	writeState writeState
	logger     *logging.Logger
}

// Create a new State object
func NewState(db dbm.DB) *State {
	cacheDB := storage.NewCacheDB(db)
	forest, err := storage.NewMutableForest(storage.NewPrefixDB(cacheDB, forestPrefix), defaultCacheCapacity)
	if err != nil {
		// This should only happen if we have negative cache capacity, which for us is a positive compile-time constant
		panic(fmt.Errorf("could not create new state because error creating MutableForest"))
	}
	ring := validator.NewRing(nil, DefaultValidatorsWindowSize)
	rs := ReadState{Forest: forest, History: ring}
	ws := writeState{forest: forest, ring: ring}
	return &State{
		db:         db,
		cacheDB:    cacheDB,
		ReadState:  rs,
		writeState: ws,
		logger:     logging.NewNoopLogger(),
	}
}

// Make genesis state from GenesisDoc and save to DB
func MakeGenesisState(db dbm.DB, genesisDoc *genesis.GenesisDoc) (*State, error) {
	s := NewState(db)

	const errHeader = "MakeGenesisState():"
	// Make accounts state tree
	for _, genAcc := range genesisDoc.Accounts {
		perm := genAcc.Permissions
		acc := &acm.Account{
			Address:     genAcc.Address,
			Balance:     genAcc.Amount,
			Permissions: perm,
		}
		err := s.writeState.UpdateAccount(acc)
		if err != nil {
			return nil, fmt.Errorf("%s %v", errHeader, err)
		}
	}
	// Make genesis validators
	err := s.writeState.MakeGenesisValidators(genesisDoc)
	if err != nil {
		return nil, fmt.Errorf("%s %v", errHeader, err)
	}
	// global permissions are saved as the 0 address
	// so they are included in the accounts tree
	globalPerms := permission.DefaultAccountPermissions
	globalPerms = genesisDoc.GlobalPermissions
	// XXX: make sure the set bits are all true
	// Without it the HasPermission() functions will fail
	globalPerms.Base.SetBit = permission.AllPermFlags

	permsAcc := &acm.Account{
		Address:     acm.GlobalPermissionsAddress,
		Balance:     1337,
		Permissions: globalPerms,
	}
	err = s.writeState.UpdateAccount(permsAcc)
	if err != nil {
		return nil, fmt.Errorf("%s %v", errHeader, err)
	}
	_, version, err := s.commit()
	if err != nil {
		return nil, fmt.Errorf("%s could not save genesis state: %v", errHeader, err)
	}
	if version != VersionOffset {
		return nil, fmt.Errorf("%s got version %d after committing genesis state but version offset should be %d",
			errHeader, version, VersionOffset)
	}
	return s, nil
}

// Tries to load the execution state from DB, returns nil with no error if no state found
func LoadState(db dbm.DB, version int64) (*State, error) {
	s := NewState(db)
	err := s.writeState.forest.Load(version)
	if err != nil {
		return nil, fmt.Errorf("could not load MutableForest at version %d: %v", version, err)
	}
	// Populate stats. If this starts taking too long, store the value rather than the full scan at startup
	err = s.IterateAccounts(func(acc *acm.Account) error {
		if len(acc.Code) > 0 {
			s.writeState.accountStats.AccountsWithCode++
		} else {
			s.writeState.accountStats.AccountsWithoutCode++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// load the validator ring
	ring, err := LoadValidatorRing(version, DefaultValidatorsWindowSize, s.writeState.forest.GetImmutable)
	if err != nil {
		return nil, err
	}
	s.writeState.ring = ring
	s.ReadState.History = ring

	return s, nil
}

func (s *State) Version() int64 {
	return s.writeState.forest.Version()
}

func (s *State) Hash() []byte {
	return s.writeState.forest.Hash()
}

func (s *State) LoadHeight(height uint64) (*ReadState, error) {
	version := VersionAtHeight(height)
	forest, err := s.writeState.forest.GetImmutable(version)
	if err != nil {
		return nil, err
	}
	ring, err := LoadValidatorRing(version, DefaultValidatorsWindowSize, s.writeState.forest.GetImmutable)
	if err != nil {
		return nil, err
	}
	return &ReadState{
		Forest:  forest,
		History: ring,
	}, nil
}

// Perform updates to state whilst holding the write lock, allows a commit to hold the write lock across multiple
// operations while preventing interlaced reads and writes
func (s *State) Update(updater func(up Updatable) error) ([]byte, int64, error) {
	s.Lock()
	defer s.Unlock()
	err := updater(&s.writeState)
	if err != nil {
		return nil, 0, err
	}
	return s.commit()
}

func (s *State) commit() ([]byte, int64, error) {
	// save state at a new version may still be orphaned before we save the version against the hash
	hash, version, err := s.writeState.forest.Save()
	if err != nil {
		return nil, 0, err
	}
	totalPowerChange, totalFlow, err := s.writeState.ring.Rotate()
	if err != nil {
		return nil, 0, err
	}
	if totalFlow.Sign() != 0 {
		//noinspection ALL
		s.logger.InfoMsg("validator set changes", "total_power_change", totalPowerChange, "total_flow", totalFlow)
	}
	// Commit the state in cacheDB atomically for this block (synchronous)
	batch := s.db.NewBatch()
	s.cacheDB.Commit(batch)
	batch.WriteSync()
	return hash, version, err
}

// Creates a copy of the database to the supplied db
func (s *State) Copy(db dbm.DB) (*State, error) {
	stateCopy := NewState(db)
	err := s.writeState.forest.IterateRWTree(nil, nil, true,
		func(prefix []byte, tree *storage.RWTree) error {
			treeCopy, err := stateCopy.writeState.forest.Writer(prefix)
			if err != nil {
				return err
			}
			return tree.IterateWriteTree(nil, nil, true, func(key []byte, value []byte) error {
				treeCopy.Set(key, value)
				return nil
			})
		})
	if err != nil {
		return nil, err
	}
	_, _, err = stateCopy.commit()
	if err != nil {
		return nil, err
	}
	return stateCopy, nil
}

func (s *State) SetLogger(logger *logging.Logger) {
	s.logger = logger
}
