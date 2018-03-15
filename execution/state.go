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

package execution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/merkleeyes/iavl"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/merkle"
)

var (
	stateKey                     = []byte("ExecutionState")
	defaultAccountsCacheCapacity = 1000 // TODO adjust
)

// TODO
const GasLimit = uint64(1000000)

//-----------------------------------------------------------------------------

// NOTE: not goroutine-safe.
type State struct {
	sync.RWMutex
	db            dbm.DB
	accounts      merkle.Tree // Shouldn't be accessed directly.
	nameReg       merkle.Tree // Shouldn't be accessed directly.
	lastSavedHash []byte
	logger        logging_types.InfoTraceLogger
}

// Implements account and blockchain state
var _ acm.Updater = &State{}
var _ acm.StateIterable = &State{}
var _ acm.StateWriter = &State{}

type PersistedState struct {
	AccountsRootHash []byte
	NameRegHash      []byte
}

func newState(db dbm.DB) *State {
	return &State{
		db:       db,
		accounts: iavl.NewIAVLTree(defaultAccountsCacheCapacity, db),
		nameReg:  iavl.NewIAVLTree(0, db),
	}
}

func LoadOrMakeGenesisState(db dbm.DB, genesisDoc *genesis.GenesisDoc,
	logger logging_types.InfoTraceLogger) (*State, error) {

	logger = logging.WithScope(logger, "LoadOrMakeGenesisState")
	logging.InfoMsg(logger, "Trying to load execution state from database",
		"database_key", stateKey)
	state, err := LoadState(db)
	if err != nil {
		return nil, fmt.Errorf("error loading genesis state from database: %v", err)
	}
	if state != nil {
		return state, nil
	}

	logging.InfoMsg(logger, "No existing execution state found in database, making genesis state")
	return MakeGenesisState(db, genesisDoc)
}

// Make genesis state from GenesisDoc and save to DB
func MakeGenesisState(db dbm.DB, genesisDoc *genesis.GenesisDoc) (*State, error) {
	if len(genesisDoc.Validators) == 0 {
		return nil, fmt.Errorf("the genesis file has no validators")
	}

	state := newState(db)

	if genesisDoc.GenesisTime.IsZero() {
		// NOTE: [ben] change GenesisTime to requirement on v0.17
		// GenesisTime needs to be deterministic across the chain
		// and should be required in the genesis file;
		// the requirement is not yet enforced when lacking set
		// time to 11/18/2016 @ 4:09am (UTC)
		genesisDoc.GenesisTime = time.Unix(1479442162, 0)
	}

	// Make accounts state tree
	for _, genAcc := range genesisDoc.Accounts {
		perm := genAcc.Permissions
		acc := &acm.ConcreteAccount{
			Address:     genAcc.Address,
			Balance:     genAcc.Amount,
			Permissions: perm,
		}
		encodedAcc, err := acc.Encode()
		if err != nil {
			return nil, err
		}
		state.accounts.Set(acc.Address.Bytes(), encodedAcc)
	}

	// global permissions are saved as the 0 address
	// so they are included in the accounts tree
	globalPerms := ptypes.DefaultAccountPermissions
	globalPerms = genesisDoc.GlobalPermissions
	// XXX: make sure the set bits are all true
	// Without it the HasPermission() functions will fail
	globalPerms.Base.SetBit = ptypes.AllPermFlags

	permsAcc := &acm.ConcreteAccount{
		Address:     permission.GlobalPermissionsAddress,
		Balance:     1337,
		Permissions: globalPerms,
	}
	encodedPermsAcc, err := permsAcc.Encode()
	if err != nil {
		return nil, err
	}
	state.accounts.Set(permsAcc.Address.Bytes(), encodedPermsAcc)

	// IAVLTrees must be persisted before copy operations.
	err = state.Save()
	if err != nil {
		return nil, err
	}
	return state, nil

}

// Tries to load the execution state from DB, returns nil with no error if no state found
func LoadState(db dbm.DB) (*State, error) {
	state := newState(db)
	buf := db.Get(stateKey)
	if len(buf) == 0 {
		return nil, nil
	}
	persistedState, err := Decode(buf)
	if err != nil {
		return nil, err
	}
	state.accounts.Load(persistedState.AccountsRootHash)
	state.nameReg.Load(persistedState.NameRegHash)
	err = state.Save()
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (s *State) Save() error {
	s.Lock()
	defer s.Unlock()
	s.accounts.Save()
	s.nameReg.Save()
	encodedState, err := s.Encode()
	if err != nil {
		return err
	}
	s.db.SetSync(stateKey, encodedState)
	s.lastSavedHash = s.hash()
	return nil
}

func (s *State) LastSavedHash() []byte {
	return s.lastSavedHash
}

func (s *State) Encode() ([]byte, error) {
	persistedState := &PersistedState{
		AccountsRootHash: s.accounts.Hash(),
		NameRegHash:      s.nameReg.Hash(),
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

// CONTRACT:
// Copy() is a cheap way to take a snapshot,
// as if State were copied by value.
// TODO [Silas]: Kill this with fire it is totally broken - there is no safe way to copy IAVLTree while sharing database
func (s *State) copy() *State {
	return &State{
		db:       s.db,
		accounts: s.accounts.Copy(),
		nameReg:  s.nameReg.Copy(),
	}
}

// Computes the state hash, also computed on save where it is returned
func (s *State) Hash() []byte {
	s.RLock()
	defer s.RUnlock()
	return s.hash()
}

// As Hash without lock
func (s *State) hash() []byte {
	return merkle.SimpleHashFromMap(map[string]interface{}{
		"Accounts":     s.accounts,
		"NameRegistry": s.nameReg,
	})
}

// Returns nil if account does not exist with given address.
func (s *State) GetAccount(address acm.Address) (acm.Account, error) {
	s.RLock()
	defer s.RUnlock()
	_, accBytes, _ := s.accounts.Get(address.Bytes())
	if accBytes == nil {
		return nil, nil
	}
	return acm.Decode(accBytes)
}

func (s *State) UpdateAccount(account acm.Account) error {
	// TODO: interop with StateCache by performing an update on the StorageRoot here if storage is dirty
	// we need `dirtyStorage map[acm.Address]bool`
	//if dirtyStorage[account] == true {
	//	s.accountStorage(account.Address())
	//	 := acm.AsMutableAccount(account).SetStorageRoot()
	//}
	s.Lock()
	defer s.Unlock()
	encodedAccount, err := account.Encode()
	if err != nil {
		return err
	}
	s.accounts.Set(account.Address().Bytes(), encodedAccount)
	return nil
}

func (s *State) RemoveAccount(address acm.Address) error {
	s.Lock()
	defer s.Unlock()
	s.accounts.Remove(address.Bytes())
	return nil
}

// This does not give a true independent copy since the underlying database is shared and any save calls all copies
// to become invalid and using them may cause panics
func (s *State) GetAccounts() merkle.Tree {
	return s.accounts.Copy()
}

func (s *State) IterateAccounts(consumer func(acm.Account) (stop bool)) (stopped bool, err error) {
	s.RLock()
	defer s.RUnlock()
	stopped = s.accounts.Iterate(func(key, value []byte) bool {
		var account acm.Account
		account, err = acm.Decode(value)
		if err != nil {
			return true
		}
		return consumer(account)
	})
	return
}

//-------------------------------------
// State.storage

func (s *State) accountStorage(address acm.Address) (merkle.Tree, error) {
	account, err := s.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("could not find account %s to access its storage", address)
	}
	return s.LoadStorage(account.StorageRoot()), nil
}

func (s *State) LoadStorage(hash []byte) merkle.Tree {
	s.RLock()
	defer s.RUnlock()
	storage := iavl.NewIAVLTree(1024, s.db)
	storage.Load(hash)
	return storage
}

func (s *State) GetStorage(address acm.Address, key binary.Word256) (binary.Word256, error) {
	s.RLock()
	defer s.RUnlock()
	storageTree, err := s.accountStorage(address)
	if err != nil {
		return binary.Zero256, err
	}
	_, value, _ := storageTree.Get(key.Bytes())
	return binary.LeftPadWord256(value), nil
}

func (s *State) SetStorage(address acm.Address, key, value binary.Word256) error {
	// TODO: not sure this actually works - loading at old hash
	s.Lock()
	defer s.Unlock()
	storageTree, err := s.accountStorage(address)
	if err != nil {
		return err
	}
	if storageTree != nil {
		storageTree.Set(key.Bytes(), value.Bytes())
	}
	return nil
}

func (s *State) IterateStorage(address acm.Address,
	consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {

	var storageTree merkle.Tree
	storageTree, err = s.accountStorage(address)
	if err != nil {
		return
	}
	stopped = storageTree.Iterate(func(key []byte, value []byte) (stop bool) {
		// Note: no left padding should occur unless there is a bug and non-words have been writte to this storage tree
		if len(key) != binary.Word256Length {
			err = fmt.Errorf("key '%X' stored for account %s is not a %v-byte word",
				key, address, binary.Word256Length)
			return true
		}
		if len(value) != binary.Word256Length {
			err = fmt.Errorf("value '%X' stored for account %s is not a %v-byte word",
				key, address, binary.Word256Length)
			return true
		}
		return consumer(binary.LeftPadWord256(key), binary.LeftPadWord256(value))
	})
	return
}

// State.storage
//-------------------------------------
// State.nameReg

var _ NameRegIterable = &State{}

func (s *State) GetNameRegEntry(name string) *NameRegEntry {
	_, valueBytes, _ := s.nameReg.Get([]byte(name))
	if valueBytes == nil {
		return nil
	}

	return DecodeNameRegEntry(valueBytes)
}

func (s *State) IterateNameRegEntries(consumer func(*NameRegEntry) (stop bool)) (stopped bool) {
	return s.nameReg.Iterate(func(key []byte, value []byte) (stop bool) {
		return consumer(DecodeNameRegEntry(value))
	})
}

func DecodeNameRegEntry(entryBytes []byte) *NameRegEntry {
	var n int
	var err error
	value := NameRegDecode(bytes.NewBuffer(entryBytes), &n, &err)
	return value.(*NameRegEntry)
}

func (s *State) UpdateNameRegEntry(entry *NameRegEntry) bool {
	w := new(bytes.Buffer)
	var n int
	var err error
	NameRegEncode(entry, w, &n, &err)
	return s.nameReg.Set([]byte(entry.Name), w.Bytes())
}

func (s *State) RemoveNameRegEntry(name string) bool {
	_, removed := s.nameReg.Remove([]byte(name))
	return removed
}

// Set the name reg tree
func (s *State) SetNameReg(nameReg merkle.Tree) {
	s.nameReg = nameReg
}
func NameRegEncode(o interface{}, w io.Writer, n *int, err *error) {
	wire.WriteBinary(o.(*NameRegEntry), w, n, err)
}

func NameRegDecode(r io.Reader, n *int, err *error) interface{} {
	return wire.ReadBinary(&NameRegEntry{}, r, txs.MaxDataLength, n, err)
}
