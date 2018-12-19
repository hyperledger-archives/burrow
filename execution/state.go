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
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hyperledger/burrow/txs"

	"github.com/hyperledger/burrow/dump"
	"github.com/hyperledger/burrow/txs/payload"

	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/tmhash"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/proposal"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/storage"
	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	defaultCacheCapacity = 1024
	uint64Length         = 8

	// Prefix under which the versioned merkle state tree resides - tracking previous versions of history
	treePrefix = "m"
	// Prefix under which all non-versioned values reside - either immutable values of references to immutable values
	// that track the current state rather than being part of the history.
	refsPrefix = "r"
)

var (
	// Directly referenced values
	accountKeyFormat  = storage.NewMustKeyFormat("a", crypto.AddressLength)
	storageKeyFormat  = storage.NewMustKeyFormat("s", crypto.AddressLength, binary.Word256Length)
	nameKeyFormat     = storage.NewMustKeyFormat("n", storage.VariadicSegmentLength)
	proposalKeyFormat = storage.NewMustKeyFormat("p", sha256.Size)

	// Keys that reference references
	blockRefKeyFormat = storage.NewMustKeyFormat("b", uint64Length)
	txRefKeyFormat    = storage.NewMustKeyFormat("t", uint64Length, uint64Length)

	// Reference keys (that do not contribute to state hash)
	// TODO: implement content-addressing of code and optionally blocks (to allow reference to block to be stored in state tree)
	//codeKeyFormat   = storage.NewMustKeyFormat("c", sha256.Size)
	//blockKeyFormat  = storage.NewMustKeyFormat("b", sha256.Size)
	txKeyFormat = storage.NewMustKeyFormat("b", txs.HashLength)
	// Binding between apphash and version stto
	commitKeyFormat = storage.NewMustKeyFormat("v", uint64Length)
)

var cdc = amino.NewCodec()

// Implements account and blockchain state
var _ state.IterableReader = &State{}
var _ names.IterableReader = &State{}
var _ Updatable = &writeState{}

type Updatable interface {
	state.Writer
	names.Writer
	proposal.Writer
	AddBlock(blockExecution *exec.BlockExecution) error
}

// Wraps state to give access to writer methods
type writeState struct {
	state *State
}

type CommitID struct {
	Hash    binary.HexBytes
	Version int64
}

func (cid CommitID) String() string {
	return fmt.Sprintf("Commit{Hash: %v, Version: %v}", cid.Hash, cid.Version)
}

// Writers to state are responsible for calling State.Lock() before calling
type State struct {
	// Last seen height from GetBlock
	height       uint64
	accountStats state.AccountStats
	// Values not reassigned
	sync.RWMutex
	StateTree
	writeState *writeState
	db         dbm.DB
	cacheDB    *storage.CacheDB
	refs       storage.KVStore
	codec      *amino.Codec
}

type StateTree struct {
	tree *storage.RWTree
}

func newStateTree(cacheDB *storage.CacheDB) StateTree {
	return StateTree{tree: storage.NewRWTree(storage.NewPrefixDB(cacheDB, treePrefix), defaultCacheCapacity)}
}

// Create a new State object
func NewState(db dbm.DB) *State {
	// We collapse all db operations into a single batch committed by save()
	cacheDB := storage.NewCacheDB(db)
	statetree := newStateTree(cacheDB)
	refs := storage.NewPrefixDB(cacheDB, refsPrefix)
	s := &State{
		db:        db,
		cacheDB:   cacheDB,
		StateTree: statetree,
		refs:      refs,
		codec:     amino.NewCodec(),
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

	permsAcc := &acm.Account{
		Address:     acm.GlobalPermissionsAddress,
		Balance:     1337,
		Permissions: globalPerms,
	}
	err := s.writeState.UpdateAccount(permsAcc)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *State) LoadDump(filename string) error {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	tx := exec.TxExecution{TxHash: make([]byte, 32)}

	for {
		var row dump.Dump

		_, err = cdc.UnmarshalBinaryLengthPrefixedReader(f, &row, 0)
		if err != nil {
			break
		}

		if row.Account != nil {
			s.writeState.UpdateAccount(row.Account)
		}
		if row.AccountStorage != nil {
			s.writeState.SetStorage(row.AccountStorage.Address, row.AccountStorage.Storage.Key, row.AccountStorage.Storage.Value)
		}
		if row.Name != nil {
			s.writeState.UpdateName(row.Name)
		}
		if row.EVMEvent != nil {
			tx.Events = append(tx.Events, &exec.Event{Log: row.EVMEvent.Event})
		}
	}

	s.writeState.AddBlock(&exec.BlockExecution{
		Height:       0,
		TxExecutions: []*exec.TxExecution{&tx},
	})

	if err == io.EOF {
		return nil
	}

	return err
}

// Tries to load the execution state from DB, returns nil with no error if no state found
func LoadState(db dbm.DB, version int64) (*State, error) {
	s := NewState(db)
	commitID, err := s.CommitID(version)
	if err != nil {
		return nil, err
	}
	if commitID.Version <= 0 {
		return nil, fmt.Errorf("trying to load state from non-positive version: CommitID: %v", commitID)
	}
	err = s.tree.Load(commitID.Version)
	if err != nil {
		return nil, fmt.Errorf("could not load current version of state tree: CommitID: %v", commitID)
	}
	// Populate stats. If this starts taking too long, store the value rather than the full scan at startup
	_, err = s.IterateAccounts(func(acc *acm.Account) (stop bool) {
		if len(acc.Code) > 0 {
			s.accountStats.AccountsWithCode++
		} else {
			s.accountStats.AccountsWithoutCode++
		}
		return
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *State) Version() int64 {
	return s.tree.Version()
}

func (s *State) CommitID(version int64) (*CommitID, error) {
	// Get the version associated with this state hash
	commitID := new(CommitID)
	err := s.codec.UnmarshalBinaryBare(s.refs.Get(commitKeyFormat.Key(version)), commitID)
	if err != nil {
		return nil, fmt.Errorf("could not decode CommitID: %v", err)
	}
	return commitID, nil
}

func (s *State) LoadHeight(height uint64) (*StateTree, error) {
	tree, err := s.tree.GetImmutableVersion(int64(height))
	if err != nil {
		return nil, err
	}

	return &StateTree{
		tree,
	}, nil
}

// Perform updates to state whilst holding the write lock, allows a commit to hold the write lock across multiple
// operations while preventing interlaced reads and writes
func (s *State) Update(updater func(up Updatable) error) ([]byte, int64, error) {
	s.Lock()
	defer s.Unlock()
	err := updater(s.writeState)
	if err != nil {
		return nil, 0, err
	}
	return s.writeState.commit()
}

func (ws *writeState) commit() ([]byte, int64, error) {
	// save state at a new version may still be orphaned before we save the version against the hash
	hash, version, err := ws.state.tree.Save()
	if err != nil {
		return nil, 0, err
	}
	if len(hash) == 0 {
		// Normalise the hash of an empty to tree to the correct hash size
		hash = make([]byte, tmhash.Size)
	}
	// Provide a reference to load this version in the future from the state hash
	commitID := CommitID{
		Hash:    hash,
		Version: version,
	}
	bs, err := ws.state.codec.MarshalBinaryBare(commitID)
	if err != nil {
		return nil, 0, fmt.Errorf("could not encode CommitID %v: %v", commitID, err)
	}
	ws.state.refs.Set(commitKeyFormat.Key(version), bs)
	// Commit the state in cacheDB atomically for this block (synchronous)
	batch := ws.state.db.NewBatch()
	ws.state.cacheDB.Commit(batch)
	batch.WriteSync()
	return hash, version, err
}

// Returns nil if account does not exist with given address.
func (s *StateTree) GetAccount(address crypto.Address) (*acm.Account, error) {
	accBytes := s.tree.Get(accountKeyFormat.Key(address))
	if accBytes == nil {
		return nil, nil
	}
	return acm.Decode(accBytes)
}

func (ws *writeState) statsAddAccount(acc *acm.Account) {
	if acc != nil {
		if len(acc.Code) > 0 {
			ws.state.accountStats.AccountsWithCode++
		} else {
			ws.state.accountStats.AccountsWithoutCode++
		}
	}
}

func (ws *writeState) statsRemoveAccount(acc *acm.Account) {
	if acc != nil {
		if len(acc.Code) > 0 {
			ws.state.accountStats.AccountsWithCode--
		} else {
			ws.state.accountStats.AccountsWithoutCode--
		}
	}
}

func (ws *writeState) UpdateAccount(account *acm.Account) error {
	if account == nil {
		return fmt.Errorf("UpdateAccount passed nil account in State")
	}
	encodedAccount, err := account.Encode()
	if err != nil {
		return fmt.Errorf("UpdateAccount could not encode account: %v", err)
	}
	updated := ws.state.tree.Set(accountKeyFormat.Key(account.Address), encodedAccount)
	if updated {
		ws.statsAddAccount(account)
	}
	return nil
}

func (ws *writeState) RemoveAccount(address crypto.Address) (err error) {
	accBytes, deleted := ws.state.tree.Delete(accountKeyFormat.Key(address))
	if deleted {
		var acc *acm.Account
		acc, err = acm.Decode(accBytes)
		if err == nil {
			ws.statsRemoveAccount(acc)
		}
	}
	return err
}

func (s *StateTree) IterateAccounts(consumer func(*acm.Account) (stop bool)) (stopped bool, err error) {
	it := accountKeyFormat.Iterator(s.tree, nil, nil)
	for it.Valid() {
		account, err := acm.Decode(it.Value())
		if err != nil {
			return true, fmt.Errorf("IterateAccounts could not decode account: %v", err)
		}
		if consumer(account) {
			return true, nil
		}
		it.Next()
	}
	return false, nil
}

func (s *State) GetAccountStats() state.AccountStats {
	return s.accountStats
}

func (s *StateTree) GetStorage(address crypto.Address, key binary.Word256) (binary.Word256, error) {
	return binary.LeftPadWord256(s.tree.Get(storageKeyFormat.Key(address, key))), nil
}

func (ws *writeState) SetStorage(address crypto.Address, key, value binary.Word256) error {
	if value == binary.Zero256 {
		ws.state.tree.Delete(storageKeyFormat.Key(address, key))
	} else {
		ws.state.tree.Set(storageKeyFormat.Key(address, key), value.Bytes())
	}
	return nil
}

func (s *StateTree) IterateStorage(address crypto.Address, consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {
	it := storageKeyFormat.Fix(address).Iterator(s.tree, nil, nil)
	for it.Valid() {
		key := it.Key()
		// Note: no left padding should occur unless there is a bug and non-words have been written to this storage tree
		if len(key) != binary.Word256Length {
			return true, fmt.Errorf("key '%X' stored for account %s is not a %v-byte word",
				key, address, binary.Word256Length)
		}
		value := it.Value()
		if len(value) != binary.Word256Length {
			return true, fmt.Errorf("value '%X' stored for account %s is not a %v-byte word",
				key, address, binary.Word256Length)
		}
		if consumer(binary.LeftPadWord256(key), binary.LeftPadWord256(value)) {
			return true, nil
		}
		it.Next()
	}
	return false, nil
}

// State.storage
//-------------------------------------
// Events

// Execution events
func (ws *writeState) AddBlock(be *exec.BlockExecution) error {
	if ws.state.height > 0 && be.Height != ws.state.height+1 {
		return fmt.Errorf("AddBlock received block for height %v but last block height was %v",
			be.Height, ws.state.height)
	}
	ws.state.height = be.Height
	// Index transactions so they can be retrieved by their TxHash
	for i, txe := range be.TxExecutions {
		ws.addTx(txe.TxHash, be.Height, uint64(i))
	}
	bs, err := be.Encode()
	if err != nil {
		return err
	}
	ws.state.refs.Set(blockRefKeyFormat.Key(be.Height), bs)
	return nil
}

func (ws *writeState) addTx(txHash []byte, height, index uint64) {
	ws.state.refs.Set(txKeyFormat.Key(txHash), txRefKeyFormat.Key(height, index))
}

func (s *State) GetTx(txHash []byte) (*exec.TxExecution, error) {
	bs := s.tree.Get(txKeyFormat.Key(txHash))
	if len(bs) == 0 {
		return nil, nil
	}
	height, index := new(uint64), new(uint64)
	txRefKeyFormat.Scan(bs, height, index)
	be, err := s.GetBlock(*height)
	if err != nil {
		return nil, fmt.Errorf("error getting block %v containing tx %X", height, txHash)
	}
	if *index < uint64(len(be.TxExecutions)) {
		return be.TxExecutions[*index], nil
	}
	return nil, fmt.Errorf("retrieved index %v in block %v for tx %X but block only contains %v TxExecutions",
		index, height, txHash, len(be.TxExecutions))
}

func (s *State) IterateTx(start, end uint64, consumer func(tx *exec.TxExecution) (stop bool)) (stopped bool, err error) {
	for height := start; height <= end; height++ {
		be, err := s.GetBlock(height)
		if err != nil {
			return false, fmt.Errorf("error getting block %v", height)
		}
		if be == nil {
			continue
		}
		for i := 0; i < len(be.TxExecutions); i++ {
			stopped = consumer(be.TxExecutions[i])
			if stopped {
				return stopped, err
			}
		}
	}
	return false, nil
}

func (s *State) GetBlock(height uint64) (*exec.BlockExecution, error) {
	bs := s.refs.Get(blockRefKeyFormat.Key(height))
	if len(bs) == 0 {
		return nil, nil
	}
	return exec.DecodeBlockExecution(bs)
}

func (s *State) GetBlocks(startHeight, endHeight uint64, consumer func(*exec.BlockExecution) (stop bool)) (stopped bool, err error) {
	kf := blockRefKeyFormat
	it := kf.Iterator(s.refs, kf.Suffix(startHeight), kf.Suffix(endHeight))
	for it.Valid() {
		block, err := exec.DecodeBlockExecution(it.Value())
		if err != nil {
			return true, fmt.Errorf("error unmarshalling ExecutionEvent in GetEvents: %v", err)
		}
		if consumer(block) {
			return true, nil
		}
		it.Next()
	}
	return false, nil
}

func (s *State) Hash() []byte {
	return s.tree.Hash()
}

// Events
//-------------------------------------
// State.nameReg

var _ names.IterableReader = &State{}

func (s *StateTree) GetName(name string) (*names.Entry, error) {
	entryBytes := s.tree.Get(nameKeyFormat.Key(name))
	if entryBytes == nil {
		return nil, nil
	}

	return names.DecodeEntry(entryBytes)
}

func (ws *writeState) UpdateName(entry *names.Entry) error {
	bs, err := entry.Encode()
	if err != nil {
		return err
	}
	ws.state.tree.Set(nameKeyFormat.Key(entry.Name), bs)
	return nil
}

func (ws *writeState) RemoveName(name string) error {
	ws.state.tree.Delete(nameKeyFormat.Key(name))
	return nil
}

func (s *StateTree) IterateNames(consumer func(*names.Entry) (stop bool)) (stopped bool, err error) {
	it := nameKeyFormat.Iterator(s.tree, nil, nil)
	for it.Valid() {
		entry, err := names.DecodeEntry(it.Value())
		if err != nil {
			return true, fmt.Errorf("State.IterateNames() could not iterate over names: %v", err)
		}
		if consumer(entry) {
			return true, nil
		}
		it.Next()
	}
	return false, nil
}

// Proposal
var _ proposal.IterableReader = &State{}

func (s *StateTree) GetProposal(proposalHash []byte) (*payload.Ballot, error) {
	bs := s.tree.Get(proposalKeyFormat.Key(proposalHash))
	if len(bs) == 0 {
		return nil, nil
	}

	return payload.DecodeBallot(bs)
}

func (ws *writeState) UpdateProposal(proposalHash []byte, p *payload.Ballot) error {
	bs, err := p.Encode()
	if err != nil {
		return err
	}

	ws.state.tree.Set(proposalKeyFormat.Key(proposalHash), bs)
	return nil
}

func (ws *writeState) RemoveProposal(proposalHash []byte) error {
	ws.state.tree.Delete(proposalKeyFormat.Key(proposalHash))
	return nil
}

func (s *StateTree) IterateProposals(consumer func(proposalHash []byte, proposal *payload.Ballot) (stop bool)) (stopped bool, err error) {
	it := proposalKeyFormat.Iterator(s.tree, nil, nil)
	for it.Valid() {
		entry, err := payload.DecodeBallot(it.Value())
		if err != nil {
			return true, fmt.Errorf("State.IterateProposal() could not iterate over proposals: %v", err)
		}
		if consumer(it.Key(), entry) {
			return true, nil
		}
		it.Next()
	}
	return false, nil
}

// Creates a copy of the database to the supplied db
func (s *State) Copy(db dbm.DB) (*State, error) {
	stateCopy := NewState(db)
	s.tree.IterateRange(nil, nil, true, func(key, value []byte) bool {
		stateCopy.tree.Set(key, value)
		return false
	})
	_, _, err := stateCopy.writeState.commit()
	if err != nil {
		return nil, err
	}
	return stateCopy, nil
}
