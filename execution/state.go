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
	"crypto/sha256"
	"fmt"
	"sync"

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
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	defaultCacheCapacity = 1024
	uint64Length         = 8

	// Prefix under which the versioned merkle state tree resides - tracking previous versions of history
	forestPrefix = "f"
)

var (
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
)

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

type ReadState struct {
	forest storage.ForestReader
}

func NewReadState(forest storage.ForestReader) ReadState {
	return ReadState{forest: forest}
}

// Writers to state are responsible for calling State.Lock() before calling
type State struct {
	sync.Mutex
	ReadState
	writeState   *writeState
	db           dbm.DB
	cacheDB      *storage.CacheDB
	forest       *storage.MutableForest
	accountStats state.AccountStats
}

// Create a new State object
func NewState(db dbm.DB) *State {
	cacheDB := storage.NewCacheDB(db)
	forest, err := storage.NewMutableForest(storage.NewPrefixDB(cacheDB, forestPrefix), defaultCacheCapacity)
	if err != nil {
		// This should only happen if we have negative cache capacity, which for us is a positive compile-time constant
		panic(fmt.Errorf("could not create new state because error creating RWForest"))
	}
	s := &State{
		db:        db,
		ReadState: NewReadState(forest),
		cacheDB:   cacheDB,
		forest:    forest,
	}

	s.writeState = &writeState{state: s}
	return s
}

// Make genesis state from GenesisDoc and save to DB
func MakeGenesisState(db dbm.DB, genesisDoc *genesis.GenesisDoc) (*State, error) {
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

	tx := exec.TxExecution{
		TxType: payload.TypeCall,
		TxHash: make([]byte, txs.HashLength),
	}

	apply := func(row dump.Dump) error {
		if row.Account != nil {
			if row.Account.Address != acm.GlobalPermissionsAddress {
				return s.writeState.UpdateAccount(row.Account)
			}
		}
		if row.AccountStorage != nil {
			for _, storage := range row.AccountStorage.Storage {
				err := s.writeState.SetStorage(row.AccountStorage.Address, storage.Key, storage.Value)
				if err != nil {
					return err
				}
			}
		}
		if row.Name != nil {
			return s.writeState.UpdateName(row.Name)
		}
		if row.EVMEvent != nil {
			tx.Events = append(tx.Events, &exec.Event{
				Header: &exec.Header{
					TxType:    payload.TypeCall,
					EventType: exec.TypeLog,
					Height:    row.Height,
				},
				Log: row.EVMEvent,
			})
		}
		return nil
	}

	// first try amino
	first := true

	for err == nil {
		var row dump.Dump

		_, err = cdc.UnmarshalBinaryLengthPrefixedReader(f, &row, 0)
		if err != nil {
			break
		}

		first = false
		err = apply(row)
	}

	// if we failed at the first row, try json
	if err != io.EOF && first {
		err = nil
		f.Seek(0, 0)

		decoder := json.NewDecoder(f)

		for err == nil {
			var row dump.Dump

			err = decoder.Decode(&row)
			if err != nil {
				break
			}

			err = apply(row)
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
	err := s.forest.Load(version)
	if err != nil {
		return nil, fmt.Errorf("could not load RWForest at version %d: %v", version, err)
	}
	// Populate stats. If this starts taking too long, store the value rather than the full scan at startup
	err = s.IterateAccounts(func(acc *acm.Account) error {
		if len(acc.Code) > 0 {
			s.accountStats.AccountsWithCode++
		} else {
			s.accountStats.AccountsWithoutCode++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *State) Version() int64 {
	return s.forest.Version()
}

func (s *State) LoadHeight(height uint64) (*ReadState, error) {
	tree, err := s.forest.GetImmutable(int64(height))
	if err != nil {
		return nil, err
	}

	return &ReadState{
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
	hash, version, err := ws.state.forest.Save()
	if err != nil {
		return nil, 0, err
	}
	// Commit the state in cacheDB atomically for this block (synchronous)
	batch := ws.state.db.NewBatch()
	ws.state.cacheDB.Commit(batch)
	batch.WriteSync()
	return hash, version, err
}

// Returns nil if account does not exist with given address.
func (s *ReadState) GetAccount(address crypto.Address) (*acm.Account, error) {
	tree, err := s.forest.Reader(accountKeyFormat.Prefix())
	if err != nil {
		return nil, err
	}
	accBytes := tree.Get(accountKeyFormat.KeyNoPrefix(address))
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
	tree, err := ws.state.forest.Writer(accountKeyFormat.Prefix())
	if err != nil {
		return err
	}
	updated := tree.Set(accountKeyFormat.KeyNoPrefix(account.Address), encodedAccount)
	if updated {
		ws.statsAddAccount(account)
	}
	return nil
}

func (ws *writeState) RemoveAccount(address crypto.Address) error {
	tree, err := ws.state.forest.Writer(accountKeyFormat.Prefix())
	if err != nil {
		return err
	}
	accBytes, deleted := tree.Delete(accountKeyFormat.KeyNoPrefix(address))
	if deleted {
		acc, err := acm.Decode(accBytes)
		if err != nil {
			return err
		}
		ws.statsRemoveAccount(acc)
		// Delete storage associated with account too
		_, err = ws.state.forest.Delete(storageKeyFormat.Key(address))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ReadState) IterateAccounts(consumer func(*acm.Account) error) error {
	tree, err := s.forest.Reader(accountKeyFormat.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		account, err := acm.Decode(value)
		if err != nil {
			return fmt.Errorf("IterateAccounts could not decode account: %v", err)
		}
		return consumer(account)
	})
}

func (s *State) GetAccountStats() state.AccountStats {
	return s.accountStats
}

func (s *ReadState) GetStorage(address crypto.Address, key binary.Word256) (binary.Word256, error) {
	keyFormat := storageKeyFormat.Fix(address)
	tree, err := s.forest.Reader(keyFormat.Prefix())
	if err != nil {
		return binary.Zero256, err
	}
	return binary.LeftPadWord256(tree.Get(keyFormat.KeyNoPrefix(key))), nil
}

func (ws *writeState) SetStorage(address crypto.Address, key, value binary.Word256) error {
	keyFormat := storageKeyFormat.Fix(address)
	tree, err := ws.state.forest.Writer(keyFormat.Prefix())
	if err != nil {
		return err
	}
	if value == binary.Zero256 {
		tree.Delete(keyFormat.KeyNoPrefix(key))
	} else {
		tree.Set(keyFormat.KeyNoPrefix(key), value.Bytes())
	}
	return nil
}

func (s *ReadState) IterateStorage(address crypto.Address, consumer func(key, value binary.Word256) error) error {
	keyFormat := storageKeyFormat.Fix(address)
	tree, err := s.forest.Reader(keyFormat.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		if len(key) != binary.Word256Length {
			return fmt.Errorf("key '%X' stored for account %s is not a %v-byte word",
				key, address, binary.Word256Length)
		}
		if len(value) != binary.Word256Length {
			return fmt.Errorf("value '%X' stored for account %s is not a %v-byte word",
				key, address, binary.Word256Length)
		}
		return consumer(binary.LeftPadWord256(key), binary.LeftPadWord256(value))
	})
}

// State.storage
//-------------------------------------
// Events

func (ws *writeState) AddEvents(evs []*exec.Event) error {
	// TODO: unwrap blocks
	return nil
}

// Execution events
func (ws *writeState) AddBlock(be *exec.BlockExecution) error {
	// Index transactions so they can be retrieved by their TxHash
	for i, txe := range be.TxExecutions {
		// When restoring a dump, events are loaded without their associated Tx, so their
		// TxHash wills be all 0.
		if bytes.Compare(txe.TxHash, make([]byte, txs.HashLength)) != 0 {
			err := ws.addTx(txe.TxHash, be.Height, uint64(i))
			if err != nil {
				return err
			}
		}
	}
	bs, err := be.Encode()
	if err != nil {
		return err
	}
	tree, err := ws.state.forest.Writer(blockRefKeyFormat.Prefix())
	if err != nil {
		return err
	}
	tree.Set(blockRefKeyFormat.KeyNoPrefix(be.Height), bs)
	return nil
}

func (ws *writeState) addTx(txHash []byte, height, index uint64) error {
	tree, err := ws.state.forest.Writer(txKeyFormat.Prefix())
	if err != nil {
		return err
	}
	tree.Set(txKeyFormat.KeyNoPrefix(txHash), txRefKeyFormat.Key(height, index))
	return nil
}

func (s *State) GetTx(txHash []byte) (*exec.TxExecution, error) {
	tree, err := s.forest.Reader(txKeyFormat.Prefix())
	if err != nil {
		return nil, err
	}
	bs := tree.Get(txKeyFormat.KeyNoPrefix(txHash))
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

func (s *State) GetBlock(height uint64) (*exec.BlockExecution, error) {
	tree, err := s.forest.Reader(blockRefKeyFormat.Prefix())
	if err != nil {
		return nil, err
	}
	bs := tree.Get(blockRefKeyFormat.Key(height))
	if len(bs) == 0 {
		return nil, nil
	}
	return exec.DecodeBlockExecution(bs)
}

func (s *State) GetBlocks(startHeight, endHeight uint64, consumer func(*exec.BlockExecution) error) error {
	tree, err := s.forest.Reader(blockRefKeyFormat.Prefix())
	if err != nil {
		return err
	}
	kf := blockRefKeyFormat
	return tree.Iterate(kf.KeyNoPrefix(startHeight), kf.KeyNoPrefix(endHeight), true,
		func(key []byte, value []byte) error {
			block, err := exec.DecodeBlockExecution(value)
			if err != nil {
				return fmt.Errorf("error unmarshalling BlockExecution in GetBlocks: %v", err)
			}
			return consumer(block)
		})
}

func (s *State) Hash() []byte {
	return s.forest.Hash()
}

// Events
//-------------------------------------
// State.nameReg

var _ names.IterableReader = &State{}

func (s *ReadState) GetName(name string) (*names.Entry, error) {
	tree, err := s.forest.Reader(nameKeyFormat.Prefix())
	if err != nil {
		return nil, err
	}
	entryBytes := tree.Get(nameKeyFormat.KeyNoPrefix(name))
	if entryBytes == nil {
		return nil, nil
	}

	return names.DecodeEntry(entryBytes)
}

func (ws *writeState) UpdateName(entry *names.Entry) error {
	tree, err := ws.state.forest.Writer(nameKeyFormat.Prefix())
	if err != nil {
		return err
	}
	bs, err := entry.Encode()
	if err != nil {
		return err
	}
	tree.Set(nameKeyFormat.KeyNoPrefix(entry.Name), bs)
	return nil
}

func (ws *writeState) RemoveName(name string) error {
	tree, err := ws.state.forest.Writer(nameKeyFormat.Prefix())
	if err != nil {
		return err
	}
	tree.Delete(nameKeyFormat.KeyNoPrefix(name))
	return nil
}

func (s *ReadState) IterateNames(consumer func(*names.Entry) error) error {
	tree, err := s.forest.Reader(nameKeyFormat.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		entry, err := names.DecodeEntry(value)
		if err != nil {
			return fmt.Errorf("State.IterateNames() could not iterate over names: %v", err)
		}
		return consumer(entry)
	})
}

// Proposal
var _ proposal.IterableReader = &State{}

func (s *ReadState) GetProposal(proposalHash []byte) (*payload.Ballot, error) {
	tree, err := s.forest.Reader(proposalKeyFormat.Prefix())
	if err != nil {
		return nil, err
	}
	bs := tree.Get(proposalKeyFormat.KeyNoPrefix(proposalHash))
	if len(bs) == 0 {
		return nil, nil
	}

	return payload.DecodeBallot(bs)
}

func (ws *writeState) UpdateProposal(proposalHash []byte, p *payload.Ballot) error {
	tree, err := ws.state.forest.Writer(proposalKeyFormat.Prefix())
	if err != nil {
		return err
	}
	bs, err := p.Encode()
	if err != nil {
		return err
	}

	tree.Set(proposalKeyFormat.KeyNoPrefix(proposalHash), bs)
	return nil
}

func (ws *writeState) RemoveProposal(proposalHash []byte) error {
	tree, err := ws.state.forest.Writer(proposalKeyFormat.Prefix())
	if err != nil {
		return err
	}
	tree.Delete(proposalKeyFormat.KeyNoPrefix(proposalHash))
	return nil
}

func (s *ReadState) IterateProposals(consumer func(proposalHash []byte, proposal *payload.Ballot) error) error {
	tree, err := s.forest.Reader(proposalKeyFormat.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		entry, err := payload.DecodeBallot(value)
		if err != nil {
			return fmt.Errorf("State.IterateProposal() could not iterate over proposals: %v", err)
		}
		return consumer(key, entry)
	})
}

// Creates a copy of the database to the supplied db
func (s *State) Copy(db dbm.DB) (*State, error) {
	stateCopy := NewState(db)
	s.forest.IterateRWTree(nil, nil, true, func(prefix []byte, tree *storage.RWTree) error {
		treeCopy, err := stateCopy.forest.Writer(prefix)
		if err != nil {
			return err
		}
		return tree.IterateWriteTree(nil, nil, true, func(key []byte, value []byte) error {
			treeCopy.Set(key, value)
			return nil
		})
	})
	_, _, err := stateCopy.writeState.commit()
	if err != nil {
		return nil, err
	}
	return stateCopy, nil
}

func (s *State) GetBlockHash(blockHeight uint64) (binary.Word256, error) {
	be, err := s.GetBlock(blockHeight)
	if err != nil {
		return binary.Zero256, err
	}
	if be == nil {
		return binary.Zero256, fmt.Errorf("block %v does not exist", blockHeight)
	}
	return binary.LeftPadWord256(be.BlockHeader.AppHash), nil
}
