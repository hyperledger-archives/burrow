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
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/util"
	"github.com/hyperledger/burrow/word"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/merkleeyes/iavl"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/merkle"
)

var (
	stateKey                     = []byte("stateKey")
	minBondAmount                = int64(1)           // TODO adjust
	defaultAccountsCacheCapacity = 1000               // TODO adjust
	unbondingPeriodBlocks        = int(60 * 24 * 365) // TODO probably better to make it time based.
	validatorTimeoutBlocks       = int(10)            // TODO adjust
	maxLoadStateElementSize      = 0                  // no max
)

// TODO
const GasLimit = int64(1000000)

//-----------------------------------------------------------------------------

// NOTE: not goroutine-safe.
type State struct {
	mtx sync.RWMutex
	db  dbm.DB
	//	BondedValidators     *types.ValidatorSet
	//	LastBondedValidators *types.ValidatorSet
	//	UnbondingValidators  *types.ValidatorSet
	accounts       merkle.Tree // Shouldn't be accessed directly.
	validatorInfos merkle.Tree // Shouldn't be accessed directly.
	nameReg        merkle.Tree // Shouldn't be accessed directly.
}

// Implements account and blockchain state
var _ acm.Updater = &State{}

var _ acm.StateIterable = &State{}

var _ acm.StateWriter = &State{}

func MakeGenesisState(db dbm.DB, genDoc *genesis.GenesisDoc) *State {
	if len(genDoc.Validators) == 0 {
		util.Fatalf("The genesis file has no validators")
	}

	if genDoc.GenesisTime.IsZero() {
		// NOTE: [ben] change GenesisTime to requirement on v0.17
		// GenesisTime needs to be deterministic across the chain
		// and should be required in the genesis file;
		// the requirement is not yet enforced when lacking set
		// time to 11/18/2016 @ 4:09am (UTC)
		genDoc.GenesisTime = time.Unix(1479442162, 0)
	}

	// Make accounts state tree
	accounts := iavl.NewIAVLTree(defaultAccountsCacheCapacity, db)
	for _, genAcc := range genDoc.Accounts {
		perm := ptypes.ZeroAccountPermissions
		if genAcc.Permissions != nil {
			perm = *genAcc.Permissions
		}
		acc := &acm.ConcreteAccount{
			Address:     genAcc.Address,
			Balance:     genAcc.Amount,
			Permissions: perm,
		}
		accounts.Set(acc.Address.Bytes(), acc.Encode())
	}

	// global permissions are saved as the 0 address
	// so they are included in the accounts tree
	globalPerms := ptypes.DefaultAccountPermissions
	if genDoc.Params != nil && genDoc.Params.GlobalPermissions != nil {
		globalPerms = *genDoc.Params.GlobalPermissions
		// XXX: make sure the set bits are all true
		// Without it the HasPermission() functions will fail
		globalPerms.Base.SetBit = ptypes.AllPermFlags
	}

	permsAcc := &acm.ConcreteAccount{
		Address:     permission.GlobalPermissionsAddress,
		Balance:     1337,
		Permissions: globalPerms,
	}
	accounts.Set(permsAcc.Address.Bytes(), permsAcc.Encode())

	// Make validatorInfos state tree && validators slice
	/*
		validatorInfos := merkle.NewIAVLTree(wire.BasicCodec, types.ValidatorInfoCodec, 0, db)
		validators := make([]*types.Validator, len(genDoc.Validators))
		for i, val := range genDoc.Validators {
			pubKey := val.PubKey
			address := pubKey.Address()

			// Make ValidatorInfo
			valInfo := &types.ValidatorInfo{
				Address:         address,
				PubKey:          pubKey,
				UnbondTo:        make([]*types.TxOutput, len(val.UnbondTo)),
				FirstBondHeight: 0,
				FirstBondAmount: val.Amount,
			}
			for i, unbondTo := range val.UnbondTo {
				valInfo.UnbondTo[i] = &types.TxOutput{
					Address: unbondTo.Address,
					Amount:  unbondTo.Amount,
				}
			}
			validatorInfos.Set(address, valInfo)

			// Make validator
			validators[i] = &types.Validator{
				Address:     address,
				PubKey:      pubKey,
				VotingPower: val.Amount,
			}
		}
	*/

	// Make namereg tree
	nameReg := iavl.NewIAVLTree(0, db)
	// TODO: add names, contracts to genesis.json

	// IAVLTrees must be persisted before copy operations.
	accounts.Save()
	//validatorInfos.Save()
	nameReg.Save()

	return &State{
		db: db,
		//BondedValidators:     types.NewValidatorSet(validators),
		//LastBondedValidators: types.NewValidatorSet(nil),
		//UnbondingValidators:  types.NewValidatorSet(nil),
		accounts: accounts,
		//validatorInfos:       validatorInfos,
		nameReg: nameReg,
	}
}

func LoadState(db dbm.DB) (*State, error) {
	s := &State{db: db}
	buf := db.Get(stateKey)
	if len(buf) == 0 {
		return nil, nil
	} else {
		r, n, err := bytes.NewReader(buf), new(int), new(error)
		wire.ReadBinaryPtr(&s, r, 0, n, err)
		if *err != nil {
			// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
			return nil, fmt.Errorf("data has been corrupted or its spec has changed: %v", *err)
		}
	}
	return s, nil
}

func (s *State) Save() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.accounts.Save()
	//s.validatorInfos.Save()
	s.nameReg.Save()
	s.db.SetSync(stateKey, wire.BinaryBytes(s))
}

// CONTRACT:
// Copy() is a cheap way to take a snapshot,
// as if State were copied by value.
// TODO [Silas]: Kill this with fire it is totally broken - there is no safe way to copy IAVLTree while sharing database
func (s *State) Copy() *State {
	return &State{
		db: s.db,
		// BondedValidators:     s.BondedValidators.Copy(),     // TODO remove need for Copy() here.
		// LastBondedValidators: s.LastBondedValidators.Copy(), // That is, make updates to the validator set
		// UnbondingValidators: s.UnbondingValidators.Copy(), // copy the valSet lazily.
		accounts: s.accounts.Copy(),
		//validatorInfos:       s.validatorInfos.Copy(),
		nameReg: s.nameReg.Copy(),
	}
}

//func (s *State) Copy() *State {
//	stateCopy := &State{
//		db:              dbm.NewMemDB(),
//		chainID:         s.chainID,
//		lastBlockHeight: s.lastBlockHeight,
//		lastBlockAppHash:   s.lastBlockAppHash,
//		lastBlockTime:   s.lastBlockTime,
//		// BondedValidators:     s.BondedValidators.Copy(),     // TODO remove need for Copy() here.
//		// LastBondedValidators: s.LastBondedValidators.Copy(), // That is, make updates to the validator set
//		// UnbondingValidators: s.UnbondingValidators.Copy(), // copy the valSet lazily.
//		accounts: copyTree(s.accounts),
//		//validatorInfos:       s.validatorInfos.Copy(),
//		nameReg: copyTree(s.nameReg),
//		evc:     nil,
//	}
//	stateCopy.Save()
//	return stateCopy
//}

// Returns a hash that represents the state data, excluding Last*
func (s *State) Hash() []byte {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return merkle.SimpleHashFromMap(map[string]interface{}{
		//"BondedValidators":    s.BondedValidators,
		//"UnbondingValidators": s.UnbondingValidators,
		"Accounts": s.accounts.Hash(),
		//"ValidatorInfos":      s.validatorInfos,
		"NameRegistry": s.nameReg,
	})
}

// Returns nil if account does not exist with given address.
func (s *State) GetAccount(address acm.Address) acm.Account {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	_, accBytes, _ := s.accounts.Get(address.Bytes())
	if accBytes == nil {
		return nil
	}
	return acm.Decode(accBytes)
}

// The account is copied before setting, so mutating it
// afterwards has no side effects.
// Implements Statelike
func (s *State) UpdateAccount(account acm.Account) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.accounts.Set(account.Address().Bytes(), account.Encode())

}

func (s *State) RemoveAccount(address acm.Address) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.accounts.Remove(address.Bytes())
}

// The returned Account is a copy, so mutating it
// has no side effects. (TODO [Silas]: Yeah you'd think, but that's bollocks, just like this merkle tree implementation)
func (s *State) GetAccounts() merkle.Tree {
	return s.accounts.Copy()
}

func (s *State) IterateAccounts(consumer func(acm.Account) (stop bool)) (stopped bool) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.accounts.Iterate(func(key, value []byte) bool {
		account := acm.Decode(value)
		// We shouldn't find an non-decodable account, but we're error-free here...
		if account != nil {
			return consumer(account)
		}
		return false
	})
}

// State.accounts
//-------------------------------------
// State.validators

// XXX: now handled by tendermint core

/*

// The returned ValidatorInfo is a copy, so mutating it
// has no side effects.
func (s *State) GetValidatorInfo(address []byte) *types.ValidatorInfo {
	_, valInfo := s.validatorInfos.Get(address)
	if valInfo == nil {
		return nil
	}
	return valInfo.(*types.ValidatorInfo).Copy()
}

// Returns false if new, true if updated.
// The valInfo is copied before setting, so mutating it
// afterwards has no side effects.
func (s *State) SetValidatorInfo(valInfo *types.ValidatorInfo) (updated bool) {
	return s.validatorInfos.Set(valInfo.Address, valInfo.Copy())
}

func (s *State) GetValidatorInfos() merkle.Tree {
	return s.validatorInfos.Copy()
}

func (s *State) unbondValidator(val *types.Validator) {
	// Move validator to UnbondingValidators
	val, removed := s.BondedValidators.Remove(val.Address)
	if !removed {
		PanicCrisis("Couldn't remove validator for unbonding")
	}
	val.UnbondHeight = s.lastBlockHeight + 1
	added := s.UnbondingValidators.Add(val)
	if !added {
		PanicCrisis("Couldn't add validator for unbonding")
	}
}

func (s *State) rebondValidator(val *types.Validator) {
	// Move validator to BondingValidators
	val, removed := s.UnbondingValidators.Remove(val.Address)
	if !removed {
		PanicCrisis("Couldn't remove validator for rebonding")
	}
	val.BondHeight = s.lastBlockHeight + 1
	added := s.BondedValidators.Add(val)
	if !added {
		PanicCrisis("Couldn't add validator for rebonding")
	}
}

func (s *State) releaseValidator(val *types.Validator) {
	// Update validatorInfo
	valInfo := s.GetValidatorInfo(val.Address)
	if valInfo == nil {
		PanicSanity("Couldn't find validatorInfo for release")
	}
	valInfo.ReleasedHeight = s.lastBlockHeight + 1
	s.SetValidatorInfo(valInfo)

	// Send coins back to UnbondTo outputs
	accounts, err := getOrMakeOutputs(s, nil, valInfo.UnbondTo)
	if err != nil {
		PanicSanity("Couldn't get or make unbondTo accounts")
	}
	adjustByOutputs(accounts, valInfo.UnbondTo)
	for _, acc := range accounts {
		s.UpdateAccount(acc)
	}

	// Remove validator from UnbondingValidators
	_, removed := s.UnbondingValidators.Remove(val.Address)
	if !removed {
		PanicCrisis("Couldn't remove validator for release")
	}
}

func (s *State) destroyValidator(val *types.Validator) {
	// Update validatorInfo
	valInfo := s.GetValidatorInfo(val.Address)
	if valInfo == nil {
		PanicSanity("Couldn't find validatorInfo for release")
	}
	valInfo.DestroyedHeight = s.lastBlockHeight + 1
	valInfo.DestroyedAmount = val.VotingPower
	s.SetValidatorInfo(valInfo)

	// Remove validator
	_, removed := s.BondedValidators.Remove(val.Address)
	if !removed {
		_, removed := s.UnbondingValidators.Remove(val.Address)
		if !removed {
			PanicCrisis("Couldn't remove validator for destruction")
		}
	}

}

// Set the validator infos tree
func (s *State) SetValidatorInfos(validatorInfos merkle.Tree) {
	s.validatorInfos = validatorInfos
}

*/

// State.validators
//-------------------------------------
// State.storage

func (s *State) accountStorage(address acm.Address) merkle.Tree {
	account := s.GetAccount(address)
	if account == nil {
		// [Silas] if we wrap the State serialisation struct in an struct that holds a logger we could log this
		// Even better we could add an error to the signature of StorageSetter
		return nil
	}
	return s.LoadStorage(account.StorageRoot())
}

func (s *State) LoadStorage(hash []byte) merkle.Tree {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	storage := iavl.NewIAVLTree(1024, s.db)
	storage.Load(hash)
	return storage
}

func (s *State) GetStorage(address acm.Address, key word.Word256) word.Word256 {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	storageTree := s.accountStorage(address)
	if storageTree != nil {
		_, value, _ := storageTree.Get(key.Bytes())
		return word.LeftPadWord256(value)
	}
	return word.Zero256
}

func (s *State) SetStorage(address acm.Address, key, value word.Word256) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	storageTree := s.accountStorage(address)
	if storageTree != nil {
		storageTree.Set(key.Bytes(), value.Bytes())
	}
}

func (s *State) IterateStorage(address acm.Address,
	consumer func(key, value word.Word256) (stop bool)) (stopped bool) {

	storageTree := s.accountStorage(address)
	if storageTree != nil {
		return storageTree.Iterate(func(key []byte, value []byte) (stop bool) {
			// Note: no left padding should occur unless someone has be writing non-words to this storage tree
			return consumer(word.LeftPadWord256(key), word.LeftPadWord256(value))
		})
	}
	return false
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
	value := NameRegCodec.Decode(bytes.NewBuffer(entryBytes), &n, &err)
	return value.(*NameRegEntry)
}

func (s *State) UpdateNameRegEntry(entry *NameRegEntry) bool {
	w := new(bytes.Buffer)
	var n int
	var err error
	NameRegCodec.Encode(entry, w, &n, &err)
	return s.nameReg.Set([]byte(entry.Name), w.Bytes())
}

func (s *State) RemoveNameRegEntry(name string) bool {
	_, removed := s.nameReg.Remove([]byte(name))
	return removed
}

func (s *State) GetNames() merkle.Tree {
	return s.nameReg.Copy()
}

// Set the name reg tree
func (s *State) SetNameReg(nameReg merkle.Tree) {
	s.nameReg = nameReg
}
func NameRegEncoder(o interface{}, w io.Writer, n *int, err *error) {
	wire.WriteBinary(o.(*NameRegEntry), w, n, err)
}

func NameRegDecoder(r io.Reader, n *int, err *error) interface{} {
	return wire.ReadBinary(&NameRegEntry{}, r, txs.MaxDataLength, n, err)
}

var NameRegCodec = wire.Codec{
	Encode: NameRegEncoder,
	Decode: NameRegDecoder,
}
