package evm

import (
	"fmt"

	"github.com/go-stack/stack"
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/permission"
)

type Interface interface {
	Reader
	Writer
	// Capture any errors when accessing or writing state - will return nil if no errors have occurred so far
	errors.Provider
	errors.Sink
	// Create a new cached state over this one inheriting any cache options
	NewCache(cacheOptions ...acmstate.CacheOption) Interface
	// Sync this state cache to into its originator
	Sync() errors.CodedError
}

type Reader interface {
	GetStorage(address crypto.Address, key binary.Word256) binary.Word256
	GetBalance(address crypto.Address) uint64
	GetPermissions(address crypto.Address) permission.AccountPermissions
	GetCode(address crypto.Address) acm.Bytecode
	GetSequence(address crypto.Address) uint64
	Exists(address crypto.Address) bool
	// GetBlockHash returns	hash of the specific block
	GetBlockHash(blockNumber uint64) (binary.Word256, error)
}

type Writer interface {
	CreateAccount(address crypto.Address)
	InitCode(address crypto.Address, code []byte)
	RemoveAccount(address crypto.Address)
	SetStorage(address crypto.Address, key, value binary.Word256)
	AddToBalance(address crypto.Address, amount uint64)
	SubtractFromBalance(address crypto.Address, amount uint64)
	SetPermission(address crypto.Address, permFlag permission.PermFlag, value bool)
	UnsetPermission(address crypto.Address, permFlag permission.PermFlag)
	AddRole(address crypto.Address, role string) bool
	RemoveRole(address crypto.Address, role string) bool
}

type State struct {
	// Where we sync
	backend acmstate.ReaderWriter
	// Block chain info
	blockchainInfo bcm.BlockchainInfo
	// Cache this State wraps
	cache *acmstate.Cache
	// Any error that may have occurred
	error errors.CodedError
	// In order for nested cache to inherit any options
	cacheOptions []acmstate.CacheOption
}

func NewState(st acmstate.ReaderWriter, bci bcm.BlockchainInfo, cacheOptions ...acmstate.CacheOption) *State {
	return &State{
		backend:      st,
		blockchainInfo: bci,
		cache:        acmstate.NewCache(st, cacheOptions...),
		cacheOptions: cacheOptions,
	}
}

func (st *State) NewCache(cacheOptions ...acmstate.CacheOption) Interface {
	return NewState(st.cache, st.blockchainInfo, append(st.cacheOptions, cacheOptions...)...)
}

func (st *State) Sync() errors.CodedError {
	// Do not sync if we have erred
	if st.error != nil {
		return st.error
	}
	err := st.cache.Sync(st.backend)
	if err != nil {
		return errors.AsException(err)
	}
	return nil
}

func (st *State) Error() errors.CodedError {
	if st.error == nil {
		return nil
	}
	return st.error
}

func (st *State) PushError(err error) {
	if st.error == nil {
		// Make sure we are not wrapping a known nil value
		ex := errors.AsException(err)
		if ex != nil {
			ex.Exception = fmt.Sprintf("%s\nStack trace: %s", ex.Exception, stack.Trace().String())
			st.error = ex
		}
	}
}

// Reader

func (st *State) GetStorage(address crypto.Address, key binary.Word256) binary.Word256 {
	value, err := st.cache.GetStorage(address, key)
	if err != nil {
		st.PushError(err)
		return binary.Zero256
	}
	return value
}

func (st *State) GetBalance(address crypto.Address) uint64 {
	acc := st.account(address)
	if acc == nil {
		return 0
	}
	return acc.Balance
}

func (st *State) GetPermissions(address crypto.Address) permission.AccountPermissions {
	acc := st.account(address)
	if acc == nil {
		return permission.AccountPermissions{}
	}
	return acc.Permissions
}

func (st *State) GetCode(address crypto.Address) acm.Bytecode {
	acc := st.account(address)
	if acc == nil {
		return nil
	}
	return acc.Code
}

func (st *State) Exists(address crypto.Address) bool {
	acc, err := st.cache.GetAccount(address)
	if err != nil {
		st.PushError(err)
		return false
	}
	if acc == nil {
		return false
	}
	return true
}

func (st *State) GetSequence(address crypto.Address) uint64 {
	acc := st.account(address)
	if acc == nil {
		return 0
	}
	return acc.Sequence
}

// Writer

func (st *State) CreateAccount(address crypto.Address) {
	if st.Exists(address) {
		st.PushError(errors.ErrorCodef(errors.ErrorCodeDuplicateAddress,
			"tried to create an account at an address that already exists: %v", address))
		return
	}
	st.updateAccount(&acm.Account{Address: address})
}

func (st *State) InitCode(address crypto.Address, code []byte) {
	acc := st.mustAccount(address)
	if acc == nil {
		st.PushError(errors.ErrorCodef(errors.ErrorCodeInvalidAddress,
			"tried to initialise code for an account that does not exist: %v", address))
		return
	}
	if acc.Code != nil {
		st.PushError(errors.ErrorCodef(errors.ErrorCodeIllegalWrite,
			"tried to initialise code for a contract that already exists: %v", address))
		return
	}
	acc.Code = code
	st.updateAccount(acc)
}

func (st *State) RemoveAccount(address crypto.Address) {
	if !st.Exists(address) {
		st.PushError(errors.ErrorCodef(errors.ErrorCodeDuplicateAddress,
			"tried to remove an account at an address that does not exist: %v", address))
		return
	}
	st.removeAccount(address)
}

func (st *State) SetStorage(address crypto.Address, key, value binary.Word256) {
	err := st.cache.SetStorage(address, key, value)
	if err != nil {
		st.PushError(err)
	}
}

func (st *State) AddToBalance(address crypto.Address, amount uint64) {
	acc := st.mustAccount(address)
	if acc == nil {
		return
	}
	if binary.IsUint64SumOverflow(acc.Balance, amount) {
		st.PushError(errors.ErrorCodef(errors.ErrorCodeIntegerOverflow,
			"uint64 overflow: attempt to add %v to the balance of %s", amount, address))
		return
	}
	acc.Balance += amount
	st.updateAccount(acc)
}

func (st *State) SubtractFromBalance(address crypto.Address, amount uint64) {
	acc := st.mustAccount(address)
	if acc == nil {
		return
	}
	if amount > acc.Balance {
		st.PushError(errors.ErrorCodef(errors.ErrorCodeInsufficientBalance,
			"insufficient funds: attempt to subtract %v from the balance of %s",
			amount, acc.Address))
		return
	}
	acc.Balance -= amount
	st.updateAccount(acc)
}

func (st *State) SetPermission(address crypto.Address, permFlag permission.PermFlag, value bool) {
	acc := st.mustAccount(address)
	if acc == nil {
		return
	}
	acc.Permissions.Base.Set(permFlag, value)
	st.updateAccount(acc)
}

func (st *State) UnsetPermission(address crypto.Address, permFlag permission.PermFlag) {
	acc := st.mustAccount(address)
	if acc == nil {
		return
	}
	acc.Permissions.Base.Unset(permFlag)
	st.updateAccount(acc)
}

func (st *State) AddRole(address crypto.Address, role string) bool {
	acc := st.mustAccount(address)
	if acc == nil {
		return false
	}
	added := acc.Permissions.AddRole(role)
	st.updateAccount(acc)
	return added
}

func (st *State) RemoveRole(address crypto.Address, role string) bool {
	acc := st.mustAccount(address)
	if acc == nil {
		return false
	}
	removed := acc.Permissions.RemoveRole(role)
	st.updateAccount(acc)
	return removed
}

// Helpers

func (st *State) account(address crypto.Address) *acm.Account {
	acc, err := st.cache.GetAccount(address)
	if err != nil {
		st.PushError(err)
	}
	return acc
}

func (st *State) mustAccount(address crypto.Address) *acm.Account {
	acc := st.account(address)
	if acc == nil {
		st.PushError(errors.ErrorCodef(errors.ErrorCodeIllegalWrite,
			"attempted to modify non-existent account: %v", address))
	}
	return acc
}

func (st *State) updateAccount(account *acm.Account) {
	err := st.cache.UpdateAccount(account)
	if err != nil {
		st.PushError(err)
	}
}

func (st *State) removeAccount(address crypto.Address) {
	err := st.cache.RemoveAccount(address)
	if err != nil {
		st.PushError(err)
	}
}

func (st *State) GetBlockHash(blockNumber uint64) (binary.Word256, error) {
	return st.blockchainInfo.GetBlockHash(blockNumber)
}
