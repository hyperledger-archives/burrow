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
	"context"
	"fmt"
	"sync"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/tendermint/iavl"
	"github.com/tendermint/tendermint/libs/pubsub"
	dbm "github.com/tendermint/tmlibs/db"
)

const (
	defaultCacheCapacity = 1024

	// Version by state hash
	versionPrefix = "v/"

	// Prefix of keys in state tree
	accountsPrefix = "a/"
	storagePrefix  = "s/"
	nameRegPrefix  = "n/"
	eventPrefix    = "e/"
)

var (
	accountsStart, accountsEnd []byte = prefixKeyRange(accountsPrefix)
	storageStart, storageEnd   []byte = prefixKeyRange(storagePrefix)
	nameRegStart, nameRegEnd   []byte = prefixKeyRange(nameRegPrefix)
)

// Implements account and blockchain state
var _ state.Iterable = &State{}
var _ names.Iterable = &State{}
var _ Updatable = &writeState{}

type Updatable interface {
	state.IterableWriter
	names.IterableWriter
	event.Publisher
	Hash() []byte
	Save() error
}

// Wraps state to give access to writer methods
type writeState struct {
	state *State
}

// Writers to state are responsible for calling State.Lock() before calling
type State struct {
	sync.RWMutex
	writeState *writeState
	// High water mark for height/index - make sure we do not overwrite events - can only increase
	eventKeyHighWatermark events.Key
	db                    dbm.DB
	tree                  *iavl.VersionedTree
	logger                *logging.Logger
}

// Create a new State object
func NewState(db dbm.DB) *State {
	s := &State{
		db:   db,
		tree: iavl.NewVersionedTree(db, defaultCacheCapacity),
	}
	s.writeState = &writeState{state: s}
	return s
}

// Make genesis state from GenesisDoc and save to DB
func MakeGenesisState(db dbm.DB, genesisDoc *genesis.GenesisDoc) (*State, error) {
	if len(genesisDoc.Validators) == 0 {
		return nil, fmt.Errorf("the genesis file has no validators")
	}

	s := NewState(db)

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
		err := s.writeState.UpdateAccount(acc.Account())
		if err != nil {
			return nil, err
		}
	}

	// global permissions are saved as the 0 address
	// so they are included in the accounts tree
	globalPerms := permission.DefaultAccountPermissions
	globalPerms = genesisDoc.GlobalPermissions
	// XXX: make sure the set bits are all true
	// Without it the HasPermission() functions will fail
	globalPerms.Base.SetBit = ptypes.AllPermFlags

	permsAcc := &acm.ConcreteAccount{
		Address:     acm.GlobalPermissionsAddress,
		Balance:     1337,
		Permissions: globalPerms,
	}
	err := s.writeState.UpdateAccount(permsAcc.Account())
	if err != nil {
		return nil, err
	}

	// IAVLTrees must be persisted before copy operations.
	err = s.writeState.Save()
	if err != nil {
		return nil, err
	}
	return s, nil

}

// Tries to load the execution state from DB, returns nil with no error if no state found
func LoadState(db dbm.DB, hash []byte) (*State, error) {
	s := NewState(db)
	// Get the version associated with this state hash
	version, err := s.writeState.GetVersion(hash)
	if err != nil {
		return nil, err
	}
	treeVersion, err := s.tree.LoadVersion(version)
	if err != nil {
		return nil, fmt.Errorf("could not load versioned state tree")
	}
	if treeVersion != version {
		return nil, fmt.Errorf("tried to load state version %v for state hash %X but loaded version %v",
			version, hash, treeVersion)
	}
	return s, nil
}

// Perform updates to state whilst holding the write lock, allows a commit to hold the write lock across multiple
// operations while preventing interlaced reads and writes
func (s *State) Update(updater func(up Updatable)) {
	s.Lock()
	defer s.Unlock()
	updater(s.writeState)
}

func (s *writeState) Save() error {
	// Save state at a new version may still be orphaned before we save the version against the hash
	hash, treeVersion, err := s.state.tree.SaveVersion()
	if err != nil {
		return err
	}
	// Provide a reference to load this version in the future from the state hash
	s.SetVersion(hash, treeVersion)
	return nil
}

// Get a previously saved tree version stored by state hash
func (s *writeState) GetVersion(hash []byte) (int64, error) {
	versionBytes := s.state.db.Get(prefixedKey(versionPrefix, hash))
	if versionBytes == nil {
		return -1, fmt.Errorf("could not retrieve version corresponding to state hash '%X' in database", hash)
	}
	return binary.GetInt64BE(versionBytes), nil
}

// Set the tree version associated with a particular hash
func (s *writeState) SetVersion(hash []byte, version int64) {
	versionBytes := make([]byte, 8)
	binary.PutInt64BE(versionBytes, version)
	s.state.db.SetSync(prefixedKey(versionPrefix, hash), versionBytes)
}

// Computes the state hash, also computed on save where it is returned
func (s *writeState) Hash() []byte {
	return s.state.tree.Hash()
}

// Returns nil if account does not exist with given address.
func (s *State) GetAccount(address crypto.Address) (acm.Account, error) {
	s.RLock()
	defer s.RUnlock()
	return s.writeState.GetAccount(address)
}

func (s *writeState) GetAccount(address crypto.Address) (acm.Account, error) {
	_, accBytes := s.state.tree.Get(prefixedKey(accountsPrefix, address.Bytes()))
	if accBytes == nil {
		return nil, nil
	}
	return acm.Decode(accBytes)
}

func (s *writeState) UpdateAccount(account acm.Account) error {
	if account == nil {
		return fmt.Errorf("UpdateAccount passed nil account in execution.State")
	}
	// TODO: find a way to implement something equivalent to this so we can set the account StorageRoot
	//storageRoot := s.tree.SubTreeHash(prefixedKey(storagePrefix, account.Address().Bytes()))
	// Alternatively just abandon and
	accountWithStorageRoot := acm.AsMutableAccount(account).SetStorageRoot(nil)
	encodedAccount, err := accountWithStorageRoot.Encode()
	if err != nil {
		return err
	}
	s.state.tree.Set(prefixedKey(accountsPrefix, account.Address().Bytes()), encodedAccount)
	return nil
}

func (s *writeState) RemoveAccount(address crypto.Address) error {
	s.state.tree.Remove(prefixedKey(accountsPrefix, address.Bytes()))
	return nil
}

func (s *State) IterateAccounts(consumer func(acm.Account) (stop bool)) (stopped bool, err error) {
	s.RLock()
	defer s.RUnlock()
	return s.writeState.IterateAccounts(consumer)
}

func (s *writeState) IterateAccounts(consumer func(acm.Account) (stop bool)) (stopped bool, err error) {
	stopped = s.state.tree.IterateRange(accountsStart, accountsEnd, true, func(key, value []byte) bool {
		var account acm.Account
		account, err = acm.Decode(value)
		if err != nil {
			return true
		}
		return consumer(account)
	})
	return
}

func (s *State) GetStorage(address crypto.Address, key binary.Word256) (binary.Word256, error) {
	s.RLock()
	defer s.RUnlock()
	return s.writeState.GetStorage(address, key)
}

func (s *writeState) GetStorage(address crypto.Address, key binary.Word256) (binary.Word256, error) {
	_, value := s.state.tree.Get(prefixedKey(storagePrefix, address.Bytes(), key.Bytes()))
	return binary.LeftPadWord256(value), nil
}

func (s *writeState) SetStorage(address crypto.Address, key, value binary.Word256) error {
	if value == binary.Zero256 {
		s.state.tree.Remove(key.Bytes())
	} else {
		s.state.tree.Set(prefixedKey(storagePrefix, address.Bytes(), key.Bytes()), value.Bytes())
	}
	return nil
}

func (s *State) IterateStorage(address crypto.Address,
	consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {
	s.RLock()
	defer s.RUnlock()
	return s.writeState.IterateStorage(address, consumer)

}

func (s *writeState) IterateStorage(address crypto.Address,
	consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {
	stopped = s.state.tree.IterateRange(storageStart, storageEnd, true, func(key []byte, value []byte) (stop bool) {
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
// Events

// Execution events
func (s *writeState) Publish(ctx context.Context, msg interface{}, tags event.Tags) error {
	if exeEvent, ok := msg.(*events.Event); ok {
		key := exeEvent.Header.Key()
		if !key.IsSuccessorOf(s.state.eventKeyHighWatermark) {
			return fmt.Errorf("received event with non-increasing key compared with current high watermark %v: %v",
				s.state.eventKeyHighWatermark, exeEvent)
		}
		s.state.eventKeyHighWatermark = key
		bs, err := exeEvent.Encode()
		if err != nil {
			return err
		}
		s.state.tree.Set(prefixedKey(eventPrefix, key), bs)
	}
	return nil
}

func (s *State) GetEvents(startBlock, endBlock uint64, queryable query.Queryable) (<-chan *events.Event, error) {
	var query pubsub.Query
	var err error
	query, err = queryable.Query()
	if err != nil {
		return nil, err
	}
	ch := make(chan *events.Event)
	go func() {
		s.RLock()
		defer s.RUnlock()
		// Close channel to signal end of iteration
		defer close(ch)
		//if endBlock == 0 {
		//endBlock = s.eventKeyHighWatermark.Height()
		//}
		s.tree.IterateRange(eventKey(startBlock, 0), eventKey(endBlock+1, 0), true,
			func(_, value []byte) bool {
				exeEvent, err := events.DecodeEvent(value)
				if err != nil {
					s.logger.InfoMsg("error unmarshalling ExecutionEvent in GetEvents", structure.ErrorKey, err)
					// stop iteration on error
					return true
				}
				if query.Matches(exeEvent) {
					ch <- exeEvent
				}
				return false
			})
	}()
	return ch, nil
}

// Events
//-------------------------------------
// State.nameReg

var _ names.Iterable = &State{}

func (s *State) GetNameEntry(name string) (*names.Entry, error) {
	s.RLock()
	defer s.RUnlock()
	return s.writeState.GetNameEntry(name)
}

func (s *writeState) GetNameEntry(name string) (*names.Entry, error) {
	_, entryBytes := s.state.tree.Get(prefixedKey(nameRegPrefix, []byte(name)))
	if entryBytes == nil {
		return nil, nil
	}

	return names.DecodeEntry(entryBytes)
}

func (s *State) IterateNameEntries(consumer func(*names.Entry) (stop bool)) (stopped bool, err error) {
	s.RLock()
	defer s.RUnlock()
	return s.writeState.IterateNameEntries(consumer)
}

func (s *writeState) IterateNameEntries(consumer func(*names.Entry) (stop bool)) (stopped bool, err error) {
	return s.state.tree.IterateRange(nameRegStart, nameRegEnd, true, func(key []byte, value []byte) (stop bool) {
		var entry *names.Entry
		entry, err = names.DecodeEntry(value)
		if err != nil {
			return true
		}
		return consumer(entry)
	}), err
}

func (s *writeState) UpdateNameEntry(entry *names.Entry) error {
	bs, err := entry.Encode()
	if err != nil {
		return err
	}
	s.state.tree.Set(prefixedKey(nameRegPrefix, []byte(entry.Name)), bs)
	return nil
}

func (s *writeState) RemoveNameEntry(name string) error {
	s.state.tree.Remove(prefixedKey(nameRegPrefix, []byte(name)))
	return nil
}

// Creates a copy of the database to the supplied db
func (s *State) Copy(db dbm.DB) *State {
	s.RLock()
	defer s.RUnlock()
	stateCopy := NewState(db)
	s.tree.Iterate(func(key []byte, value []byte) bool {
		stateCopy.tree.Set(key, value)
		return false
	})
	return stateCopy
}

func eventKey(height, index uint64) events.Key {
	return prefixedKey(eventPrefix, events.NewKey(height, index))
}

func prefixedKey(prefix string, suffices ...[]byte) []byte {
	key := []byte(prefix)
	for _, suffix := range suffices {
		key = append(key, suffix...)
	}
	return key
}

// Returns the start key equal to the bytes of prefix and the end key which lexicographically above any key beginning
// with prefix
func prefixKeyRange(prefix string) (start, end []byte) {
	start = []byte(prefix)
	for i := len(start) - 1; i >= 0; i-- {
		c := start[i]
		if c < 0xff {
			end = make([]byte, i+1)
			copy(end, start)
			end[i]++
			return
		}
	}
	return
}
