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
	"time"

	acm "github.com/hyperledger/burrow/account"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/go-wire"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/events"

	"sync"

	core_types "github.com/hyperledger/burrow/core/types"
	"github.com/hyperledger/burrow/util"
	"github.com/tendermint/merkleeyes/iavl"
	"github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/merkle"
	"github.com/hyperledger/burrow/genesis"
)

var (
	stateKey                     = []byte("stateKey")
	minBondAmount                = int64(1)           // TODO adjust
	defaultAccountsCacheCapacity = 1000               // TODO adjust
	unbondingPeriodBlocks        = int(60 * 24 * 365) // TODO probably better to make it time based.
	validatorTimeoutBlocks       = int(10)            // TODO adjust
	maxLoadStateElementSize      = 0                  // no max
)

//-----------------------------------------------------------------------------

// NOTE: not goroutine-safe.
type State struct {
	mtx             sync.Mutex
	db              dbm.DB
	ChainID         string
	LastBlockHeight uint64
	LastBlockHash   []byte
	LastBlockTime   time.Time
	// AppHash is updated after Commit
	LastBlockAppHash []byte
	//	BondedValidators     *types.ValidatorSet
	//	LastBondedValidators *types.ValidatorSet
	//	UnbondingValidators  *types.ValidatorSet
	accounts       merkle.Tree // Shouldn't be accessed directly.
	validatorInfos merkle.Tree // Shouldn't be accessed directly.
	nameReg        merkle.Tree // Shouldn't be accessed directly.

	evc events.Fireable // typically an events.EventCache
}

func LoadState(db dbm.DB) *State {
	s := &State{db: db}
	buf := db.Get(stateKey)
	if len(buf) == 0 {
		return nil
	} else {
		r, n, err := bytes.NewReader(buf), new(int), new(error)
		wire.ReadBinaryPtr(&s, r, 0, n, err)
		if *err != nil {
			// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
			util.Fatalf("Data has been corrupted or its spec has changed: %v\n", *err)
		}
		// TODO: ensure that buf is completely read.
	}
	return s
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
func (s *State) Copy() *State {
	return &State{
		db:               s.db,
		ChainID:          s.ChainID,
		LastBlockHeight:  s.LastBlockHeight,
		LastBlockHash:    s.LastBlockHash,
		LastBlockTime:    s.LastBlockTime,
		LastBlockAppHash: s.LastBlockAppHash,
		// BondedValidators:     s.BondedValidators.Copy(),     // TODO remove need for Copy() here.
		// LastBondedValidators: s.LastBondedValidators.Copy(), // That is, make updates to the validator set
		// UnbondingValidators: s.UnbondingValidators.Copy(), // copy the valSet lazily.
		accounts: s.accounts.Copy(),
		//validatorInfos:       s.validatorInfos.Copy(),
		nameReg: s.nameReg.Copy(),
		evc:     nil,
	}
}

// Returns a hash that represents the state data, excluding Last*
func (s *State) Hash() []byte {
	return merkle.SimpleHashFromMap(map[string]interface{}{
		//"BondedValidators":    s.BondedValidators,
		//"UnbondingValidators": s.UnbondingValidators,
		"Accounts": s.accounts,
		//"ValidatorInfos":      s.validatorInfos,
		"NameRegistry": s.nameReg,
		"AppHash":      s.LastBlockAppHash,
	})
}

func (s *State) GetGenesisDoc() (*types.GenesisDoc, error) {
	var genesisDoc *types.GenesisDoc
	loadedGenesisDocBytes := s.db.Get(types.GenDocKey)
	err := new(error)
	wire.ReadJSONPtr(&genesisDoc, loadedGenesisDocBytes, err)
	if *err != nil {
		return nil, fmt.Errorf("Unable to read genesisDoc from db on Get: %v", err)
	}
	return genesisDoc, nil
}

func (s *State) SetDB(db dbm.DB) {
	s.db = db
}

//-------------------------------------
// State.params

func (s *State) GetGasLimit() int64 {
	return 1000000 // TODO
}

// State.params
//-------------------------------------
// State.accounts

// Returns nil if account does not exist with given address.
// Implements Statelike
func (s *State) GetAccount(address []byte) *acm.Account {
	_, accBytes, _ := s.accounts.Get(address)
	if accBytes == nil {
		return nil
	}
	return acm.DecodeAccount(accBytes)
}

// The account is copied before setting, so mutating it
// afterwards has no side effects.
// Implements Statelike
func (s *State) UpdateAccount(account *acm.Account) bool {
	return s.accounts.Set(account.Address, acm.EncodeAccount(account))
}

// Implements Statelike
func (s *State) RemoveAccount(address []byte) bool {
	_, removed := s.accounts.Remove(address)
	return removed
}

// The returned Account is a copy, so mutating it
// has no side effects.
func (s *State) GetAccounts() merkle.Tree {
	return s.accounts.Copy()
}

// Set the accounts tree
func (s *State) SetAccounts(accounts merkle.Tree) {
	s.accounts = accounts
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
	val.UnbondHeight = s.LastBlockHeight + 1
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
	val.BondHeight = s.LastBlockHeight + 1
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
	valInfo.ReleasedHeight = s.LastBlockHeight + 1
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
	valInfo.DestroyedHeight = s.LastBlockHeight + 1
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

func (s *State) LoadStorage(hash []byte) (storage merkle.Tree) {
	storage = iavl.NewIAVLTree(1024, s.db)
	storage.Load(hash)
	return storage
}

// State.storage
//-------------------------------------
// State.nameReg

func (s *State) GetNameRegEntry(name string) *core_types.NameRegEntry {
	_, valueBytes, _ := s.nameReg.Get([]byte(name))
	if valueBytes == nil {
		return nil
	}

	return DecodeNameRegEntry(valueBytes)
}

func DecodeNameRegEntry(entryBytes []byte) *core_types.NameRegEntry {
	var n int
	var err error
	value := NameRegCodec.Decode(bytes.NewBuffer(entryBytes), &n, &err)
	return value.(*core_types.NameRegEntry)
}

func (s *State) UpdateNameRegEntry(entry *core_types.NameRegEntry) bool {
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
	wire.WriteBinary(o.(*core_types.NameRegEntry), w, n, err)
}

func NameRegDecoder(r io.Reader, n *int, err *error) interface{} {
	return wire.ReadBinary(&core_types.NameRegEntry{}, r, txs.MaxDataLength, n, err)
}

var NameRegCodec = wire.Codec{
	Encode: NameRegEncoder,
	Decode: NameRegDecoder,
}

// State.nameReg
//-------------------------------------

// Implements events.Eventable. Typically uses events.EventCache
func (s *State) SetFireable(evc events.Fireable) {
	s.evc = evc
}

//-----------------------------------------------------------------------------
// Genesis

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
		acc := &acm.Account{
			Address:     genAcc.Address,
			Balance:     genAcc.Amount,
			Permissions: perm,
		}
		accounts.Set(acc.Address, acm.EncodeAccount(acc))
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

	permsAcc := &acm.Account{
		Address:     ptypes.GlobalPermissionsAddress,
		Balance:     1337,
		Permissions: globalPerms,
	}
	accounts.Set(permsAcc.Address, acm.EncodeAccount(permsAcc))

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
		db:            db,
		ChainID:       genDoc.ChainID,
		LastBlockTime: genDoc.GenesisTime,
		//BondedValidators:     types.NewValidatorSet(validators),
		//LastBondedValidators: types.NewValidatorSet(nil),
		//UnbondingValidators:  types.NewValidatorSet(nil),
		accounts: accounts,
		//validatorInfos:       validatorInfos,
		nameReg: nameReg,
	}
}
