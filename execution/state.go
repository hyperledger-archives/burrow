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
	"fmt"
	"io"
	"sync"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/iavl"
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
)

var (
	accountsStart, accountsEnd []byte = prefixKeyRange(accountsPrefix)
	storageStart, storageEnd   []byte = prefixKeyRange(storagePrefix)
	nameRegStart, nameRegEnd   []byte = prefixKeyRange(nameRegPrefix)
)

// Implements account and blockchain state
var _ state.AccountUpdater = &State{}
var _ state.Iterable = &State{}
var _ state.Writer = &State{}

type State struct {
	sync.RWMutex
	db      dbm.DB
	version uint64
	// TODO:
	tree   *iavl.VersionedTree
	logger *logging.Logger
}

func NewState(db dbm.DB) *State {
	return &State{
		db:   db,
		tree: iavl.NewVersionedTree(db, defaultCacheCapacity),
	}
}

// Make genesis state from GenesisDoc and save to DB
func MakeGenesisState(db dbm.DB, genesisDoc *genesis.GenesisDoc) (*State, error) {
	if len(genesisDoc.Validators) == 0 {
		return nil, fmt.Errorf("the genesis file has no validators")
	}

	state := NewState(db)

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
		err := state.UpdateAccount(acc.Account())
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
	err := state.UpdateAccount(permsAcc.Account())
	if err != nil {
		return nil, err
	}

	// IAVLTrees must be persisted before copy operations.
	err = state.Save()
	if err != nil {
		return nil, err
	}
	return state, nil

}

// Tries to load the execution state from DB, returns nil with no error if no state found
func LoadState(db dbm.DB, hash []byte) (*State, error) {
	versionBytes := db.Get(prefixedKey(versionPrefix, hash))
	if versionBytes == nil {
		return nil, fmt.Errorf("could not retrieve version corresponding to state hash '%X' in database", hash)
	}
	s := NewState(db)
	s.version = binary.GetUint64BE(versionBytes)
	treeVersion, err := s.tree.Load()
	if err != nil {
		return nil, fmt.Errorf("could not load versioned state tree")
	}

	if uint64(treeVersion) != s.version {
		return nil, fmt.Errorf("LoadState expects tree version %v for state hash %X but latest state tree version "+
			"loaded is %v", s.version, hash, treeVersion)
	}
	return s, nil
}

func (s *State) Save() error {
	s.Lock()
	defer s.Unlock()
	s.version++
	hash, treeVersion, err := s.tree.SaveVersion()
	if err != nil {
		return err
	}
	if uint64(treeVersion) != s.version {
		return fmt.Errorf("Save expects state tree version %v for state hash %X tree saved as version %v",
			s.version, hash, treeVersion)
	}
	versionBytes := make([]byte, 8)
	binary.PutUint64BE(versionBytes, s.version)
	s.db.SetSync(prefixedKey(versionPrefix, hash), versionBytes)
	return nil
}

// Computes the state hash, also computed on save where it is returned
func (s *State) Hash() []byte {
	s.RLock()
	defer s.RUnlock()
	return s.tree.Hash()
}

// Returns nil if account does not exist with given address.
func (s *State) GetAccount(address crypto.Address) (acm.Account, error) {
	s.RLock()
	defer s.RUnlock()
	_, accBytes := s.tree.Get(prefixedKey(accountsPrefix, address.Bytes()))
	if accBytes == nil {
		return nil, nil
	}
	return acm.Decode(accBytes)
}

func (s *State) UpdateAccount(account acm.Account) error {
	s.Lock()
	defer s.Unlock()
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
	s.tree.Set(prefixedKey(accountsPrefix, account.Address().Bytes()), encodedAccount)
	return nil
}

func (s *State) RemoveAccount(address crypto.Address) error {
	s.Lock()
	defer s.Unlock()
	s.tree.Remove(prefixedKey(accountsPrefix, address.Bytes()))
	return nil
}

func (s *State) IterateAccounts(consumer func(acm.Account) (stop bool)) (stopped bool, err error) {
	s.RLock()
	defer s.RUnlock()
	stopped = s.tree.IterateRange(accountsStart, accountsEnd, true, func(key, value []byte) bool {
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
	_, value := s.tree.Get(prefixedKey(storagePrefix, address.Bytes(), key.Bytes()))
	return binary.LeftPadWord256(value), nil
}

func (s *State) SetStorage(address crypto.Address, key, value binary.Word256) error {
	s.Lock()
	defer s.Unlock()
	if value == binary.Zero256 {
		s.tree.Remove(key.Bytes())
	} else {
		s.tree.Set(prefixedKey(storagePrefix, address.Bytes(), key.Bytes()), value.Bytes())
	}
	return nil
}

func (s *State) IterateStorage(address crypto.Address,
	consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {
	s.RLock()
	defer s.RUnlock()

	stopped = s.tree.IterateRange(storageStart, storageEnd, true, func(key []byte, value []byte) (stop bool) {
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

var _ names.Iterable = &State{}

func (s *State) GetNameEntry(name string) (*names.Entry, error) {
	_, valueBytes := s.tree.Get(prefixedKey(nameRegPrefix, []byte(name)))
	if valueBytes == nil {
		return nil, nil
	}

	return DecodeNameRegEntry(valueBytes), nil
}

func (s *State) IterateNameEntries(consumer func(*names.Entry) (stop bool)) (stopped bool, err error) {
	return s.tree.IterateRange(nameRegStart, nameRegEnd, true, func(key []byte, value []byte) (stop bool) {
		return consumer(DecodeNameRegEntry(value))
	}), nil
}

func (s *State) UpdateNameEntry(entry *names.Entry) error {
	w := new(bytes.Buffer)
	var n int
	var err error
	NameRegEncode(entry, w, &n, &err)
	if err != nil {
		return err
	}
	s.tree.Set(prefixedKey(nameRegPrefix, []byte(entry.Name)), w.Bytes())
	return nil
}

func (s *State) RemoveNameEntry(name string) error {
	s.tree.Remove(prefixedKey(nameRegPrefix, []byte(name)))
	return nil
}

// Creates a copy of the database to the supplied db
func (s *State) Copy(db dbm.DB) *State {
	state := NewState(db)
	s.tree.Iterate(func(key []byte, value []byte) bool {
		state.tree.Set(key, value)
		return false
	})
	return state
}

func DecodeNameRegEntry(entryBytes []byte) *names.Entry {
	var n int
	var err error
	value := NameRegDecode(bytes.NewBuffer(entryBytes), &n, &err)
	return value.(*names.Entry)
}

func NameRegEncode(o interface{}, w io.Writer, n *int, err *error) {
	wire.WriteBinary(o.(*names.Entry), w, n, err)
}

func NameRegDecode(r io.Reader, n *int, err *error) interface{} {
	return wire.ReadBinary(&names.Entry{}, r, names.MaxDataLength, n, err)
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
