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
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	defaultCacheCapacity = 1024
	// Age of state versions in blocks before we remove them. This has us keeping a little over an hour's worth of blocks
	// in principle we could manage with 2. Ideally we would lift this limit altogether but IAVL leaks memory on access
	// to previous tree versions since it lazy loads values (nice) but gives no ability to unload them (see SaveBranch)
	defaultVersionExpiry = 2048

	// Version by state hash
	versionPrefix = "v/"

	// Prefix of keys in state tree
	accountsPrefix = "a/"
	storagePrefix  = "s/"
	nameRegPrefix  = "n/"
	blockPrefix    = "b/"
	txPrefix       = "t/"
)

var (
	accountsStart, accountsEnd []byte = prefixKeyRange(accountsPrefix)
	storageStart, storageEnd   []byte = prefixKeyRange(storagePrefix)
	nameRegStart, nameRegEnd   []byte = prefixKeyRange(nameRegPrefix)
	lastBlockHeightKey                = []byte("h")
)

// Implements account and blockchain state
var _ state.IterableReader = &State{}
var _ names.IterableReader = &State{}
var _ Updatable = &writeState{}

type Updatable interface {
	state.Writer
	names.Writer
	AddBlock(blockExecution *exec.BlockExecution) error
}

// Wraps state to give access to writer methods
type writeState struct {
	state *State
}

// Writers to state are responsible for calling State.Lock() before calling
type State struct {
	// Values not reassigned
	sync.RWMutex
	writeState *writeState
	db         dbm.DB
	tree       *iavl.MutableTree
	logger     *logging.Logger

	// Values may be reassigned (mutex protected)
	// Previous version of IAVL tree for concurrent read-only access
	readTree *iavl.ImmutableTree
	// Last state hash
	hash []byte
}

// Create a new State object
func NewState(db dbm.DB) *State {
	s := &State{
		db:   db,
		tree: iavl.NewMutableTree(db, defaultCacheCapacity),
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
	globalPerms.Base.SetBit = permission.AllPermFlags

	permsAcc := &acm.ConcreteAccount{
		Address:     acm.GlobalPermissionsAddress,
		Balance:     1337,
		Permissions: globalPerms,
	}
	err := s.writeState.UpdateAccount(permsAcc.Account())
	if err != nil {
		return nil, err
	}

	// We need to save at least once so that readTree points at a non-working-state tree
	_, err = s.writeState.save()
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
	if version <= 0 {
		return nil, fmt.Errorf("trying to load state from non-positive version: version %v, hash: %X", version, hash)
	}
	treeVersion, err := s.tree.LoadVersion(version)
	if err != nil {
		return nil, fmt.Errorf("could not load current version of state tree: version %v, hash: %X", version, hash)
	}
	if treeVersion != version {
		return nil, fmt.Errorf("tried to load state version %v for state hash %X but loaded version %v",
			version, hash, treeVersion)
	}
	// Load previous version for readTree
	// Set readTree
	s.readTree, err = s.tree.GetImmutable(version - 1)
	return s, nil
}

// Perform updates to state whilst holding the write lock, allows a commit to hold the write lock across multiple
// operations while preventing interlaced reads and writes
func (s *State) Update(updater func(up Updatable) error) ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	err := updater(s.writeState)
	if err != nil {
		return nil, err
	}
	return s.writeState.save()
}

func (ws *writeState) save() ([]byte, error) {
	// save state at a new version may still be orphaned before we save the version against the hash
	hash, treeVersion, err := ws.state.tree.SaveVersion()
	if err != nil {
		return nil, err
	}
	// Take an immutable reference to the tree we just saved for querying
	ws.state.readTree, err = ws.state.tree.GetImmutable(treeVersion)
	if err != nil {
		return nil, err
	}

	// Provide a reference to load this version in the future from the state hash
	ws.SetVersion(hash, treeVersion)
	ws.state.hash = hash
	return hash, err
}

// Get a previously saved tree version stored by state hash
func (ws *writeState) GetVersion(hash []byte) (int64, error) {
	versionBytes := ws.state.db.Get(prefixedKey(versionPrefix, hash))
	if versionBytes == nil {
		return -1, fmt.Errorf("could not retrieve version corresponding to state hash '%X' in database", hash)
	}
	return binary.GetInt64BE(versionBytes), nil
}

// Set the tree version associated with a particular hash
func (ws *writeState) SetVersion(hash []byte, version int64) {
	versionBytes := make([]byte, 8)
	binary.PutInt64BE(versionBytes, version)
	ws.state.db.SetSync(prefixedKey(versionPrefix, hash), versionBytes)
}

// Returns nil if account does not exist with given address.
func (s *State) GetAccount(address crypto.Address) (acm.Account, error) {
	_, accBytes := s.readTree.Get(prefixedKey(accountsPrefix, address.Bytes()))
	if accBytes == nil {
		return nil, nil
	}
	return acm.Decode(accBytes)
}

func (ws *writeState) UpdateAccount(account acm.Account) error {
	if account == nil {
		return fmt.Errorf("UpdateAccount passed nil account in State")
	}
	// TODO: find a way to implement something equivalent to this so we can set the account StorageRoot
	//storageRoot := s.tree.SubTreeHash(prefixedKey(storagePrefix, account.Address().Bytes()))
	// Alternatively just abandon and
	accountWithStorageRoot := acm.AsMutableAccount(account)
	encodedAccount, err := accountWithStorageRoot.Encode()
	if err != nil {
		return err
	}
	ws.state.tree.Set(prefixedKey(accountsPrefix, account.Address().Bytes()), encodedAccount)
	return nil
}

func (ws *writeState) RemoveAccount(address crypto.Address) error {
	ws.state.tree.Remove(prefixedKey(accountsPrefix, address.Bytes()))
	return nil
}

func (s *State) IterateAccounts(consumer func(acm.Account) (stop bool)) (stopped bool, err error) {
	stopped = s.readTree.IterateRange(accountsStart, accountsEnd, true, func(key, value []byte) bool {
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
	_, value := s.readTree.Get(prefixedKey(storagePrefix, address.Bytes(), key.Bytes()))
	return binary.LeftPadWord256(value), nil
}

func (ws *writeState) SetStorage(address crypto.Address, key, value binary.Word256) error {
	if value == binary.Zero256 {
		ws.state.tree.Remove(prefixedKey(storagePrefix, address.Bytes(), key.Bytes()))
	} else {
		ws.state.tree.Set(prefixedKey(storagePrefix, address.Bytes(), key.Bytes()), value.Bytes())
	}
	return nil
}

func (s *State) IterateStorage(address crypto.Address,
	consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {
	stopped = s.readTree.IterateRange(storageStart, storageEnd, true, func(key []byte, value []byte) (stop bool) {
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
func (ws *writeState) AddBlock(be *exec.BlockExecution) error {
	lastBlockHeight, ok := ws.lastBlockHeight()
	if ok && be.Height != lastBlockHeight+1 {
		return fmt.Errorf("AddBlock received block for height %v but last block height was %v",
			be.Height, lastBlockHeight)
	}
	ws.setLastBlockHeight(be.Height)
	// Index transactions so they can be retrieved by their TxHash
	for i, txe := range be.TxExecutions {
		ws.addTx(txe.TxHash, be.Height, uint64(i))
	}
	bs, err := be.Encode()
	if err != nil {
		return err
	}
	key := blockKey(be.Height)
	ws.state.tree.Set(key, bs)
	return nil
}

func (ws *writeState) addTx(txHash []byte, height, index uint64) {
	ws.state.tree.Set(txKey(txHash), encodeTxRef(height, index))
}

func (s *State) GetTx(txHash []byte) (*exec.TxExecution, error) {
	_, bs := s.readTree.Get(txKey(txHash))
	if len(bs) == 0 {
		return nil, nil
	}
	height, index, err := decodeTxRef(bs)
	if err != nil {
		return nil, fmt.Errorf("error decoding database reference to tx %X: %v", txHash, err)
	}
	be, err := s.GetBlock(height)
	if err != nil {
		return nil, fmt.Errorf("error getting block %v containing tx %X", height, txHash)
	}
	if index < uint64(len(be.TxExecutions)) {
		return be.TxExecutions[index], nil
	}
	return nil, fmt.Errorf("retrieved index %v in block %v for tx %X but block only contains %v TxExecutions",
		index, height, txHash, len(be.TxExecutions))
}

func (s *State) GetBlock(height uint64) (*exec.BlockExecution, error) {
	_, bs := s.readTree.Get(blockKey(height))
	if len(bs) == 0 {
		return nil, nil
	}
	return exec.DecodeBlockExecution(bs)
}

func (s *State) GetBlocks(startHeight, endHeight uint64, consumer func(*exec.BlockExecution) (stop bool)) (stopped bool, err error) {
	return s.readTree.IterateRange(blockKey(startHeight), blockKey(endHeight), true,
		func(_, value []byte) bool {
			block, err := exec.DecodeBlockExecution(value)
			if err != nil {
				err = fmt.Errorf("error unmarshalling ExecutionEvent in GetEvents: %v", err)
				// stop iteration on error
				return true
			}
			return consumer(block)
		}), err
}

func (s *State) Hash() []byte {
	s.RLock()
	defer s.RUnlock()
	return s.hash
}

func (s *writeState) lastBlockHeight() (uint64, bool) {
	_, bs := s.state.tree.Get(lastBlockHeightKey)
	if len(bs) == 0 {
		return 0, false
	}
	return binary.GetUint64BE(bs), true
}

func (s *writeState) setLastBlockHeight(height uint64) {
	bs := make([]byte, 8)
	binary.PutUint64BE(bs, height)
	s.state.tree.Set(lastBlockHeightKey, bs)
}

// Events
//-------------------------------------
// State.nameReg

var _ names.IterableReader = &State{}

func (s *State) GetName(name string) (*names.Entry, error) {
	_, entryBytes := s.readTree.Get(prefixedKey(nameRegPrefix, []byte(name)))
	if entryBytes == nil {
		return nil, nil
	}

	return names.DecodeEntry(entryBytes)
}

func (s *State) IterateNames(consumer func(*names.Entry) (stop bool)) (stopped bool, err error) {
	return s.readTree.IterateRange(nameRegStart, nameRegEnd, true, func(key []byte, value []byte) (stop bool) {
		var entry *names.Entry
		entry, err = names.DecodeEntry(value)
		if err != nil {
			return true
		}
		return consumer(entry)
	}), err
}

func (ws *writeState) UpdateName(entry *names.Entry) error {
	bs, err := entry.Encode()
	if err != nil {
		return err
	}
	ws.state.tree.Set(prefixedKey(nameRegPrefix, []byte(entry.Name)), bs)
	return nil
}

func (ws *writeState) RemoveName(name string) error {
	ws.state.tree.Remove(prefixedKey(nameRegPrefix, []byte(name)))
	return nil
}

// Creates a copy of the database to the supplied db
func (s *State) Copy(db dbm.DB) (*State, error) {
	stateCopy := NewState(db)
	s.tree.Iterate(func(key []byte, value []byte) bool {
		stateCopy.tree.Set(key, value)
		return false
	})
	_, err := stateCopy.writeState.save()
	if err != nil {
		return nil, err
	}
	return stateCopy, nil
}

// Key and value helpers

func encodeTxRef(height, index uint64) []byte {
	bs := make([]byte, 16)
	binary.PutUint64BE(bs[:8], height)
	binary.PutUint64BE(bs[8:], index)
	return bs
}

func decodeTxRef(bs []byte) (height, index uint64, _ error) {
	if len(bs) != 16 {
		return 0, 0, fmt.Errorf("tx reference must have 16 bytes but '%X' does not", bs)
	}
	height = binary.GetUint64BE(bs[:8])
	index = binary.GetUint64BE(bs[8:])
	return
}

func txKey(txHash []byte) []byte {
	return prefixedKey(txPrefix, txHash)
}

func blockKey(height uint64) []byte {
	bs := make([]byte, 8)
	binary.PutUint64BE(bs, height)
	return prefixedKey(blockPrefix, bs)
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
