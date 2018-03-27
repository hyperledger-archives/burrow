package state

import (
	"fmt"
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateCache_GetAccount(t *testing.T) {
	acc := acm.NewConcreteAccountFromSecret("foo")
	acc.Permissions.Base.Perms = permission.AddRole | permission.Send
	acc.Permissions.Base.SetBit = acc.Permissions.Base.Perms
	state := combine(account(acc.Account(), "I AM A KEY", "NO YOU ARE A KEY"))
	cache := NewCache(state)

	accOut, err := cache.GetAccount(acc.Address)
	require.NoError(t, err)
	cache.UpdateAccount(accOut)
	accEnc, err := acc.Encode()
	accEncOut, err := accOut.Encode()
	assert.Equal(t, accEnc, accEncOut)
	assert.Equal(t, acc.Permissions, accOut.Permissions())

	cacheBackend := NewCache(newTestState())
	err = cache.Sync(cacheBackend)
	require.NoError(t, err)
	accOut, err = cacheBackend.GetAccount(acc.Address)
	require.NotNil(t, accOut)
	accEncOut, err = accOut.Encode()
	assert.NoError(t, err)
	assert.Equal(t, accEnc, accEncOut)
}

func TestStateCache_UpdateAccount(t *testing.T) {
}

func TestStateCache_RemoveAccount(t *testing.T) {
}

func TestStateCache_GetStorage(t *testing.T) {
}

func TestStateCache_SetStorage(t *testing.T) {
}

func TestStateCache_Sync(t *testing.T) {
}

func TestStateCache_get(t *testing.T) {
}

// TODO: write tests as part of feature branch
type testState struct {
	Accounts map[acm.Address]acm.Account
	Storage  map[acm.Address]map[binary.Word256]binary.Word256
}

var _ Iterable = &testState{}

func newTestState() *testState {
	return &testState{
		Accounts: make(map[acm.Address]acm.Account),
		Storage:  make(map[acm.Address]map[binary.Word256]binary.Word256),
	}
}

func account(acc acm.Account, keyvals ...string) *testState {
	ts := newTestState()
	ts.Accounts[acc.Address()] = acc
	ts.Storage[acc.Address()] = make(map[binary.Word256]binary.Word256)
	for i := 0; i < len(keyvals); i += 2 {
		ts.Storage[acc.Address()][word(keyvals[i])] = word(keyvals[i+1])
	}
	return ts
}

func combine(states ...*testState) *testState {
	ts := newTestState()
	for _, state := range states {
		for _, acc := range state.Accounts {
			ts.Accounts[acc.Address()] = acc
			ts.Storage[acc.Address()] = state.Storage[acc.Address()]
		}
	}
	return ts
}

func word(str string) binary.Word256 {
	return binary.LeftPadWord256([]byte(str))
}

func (tsr *testState) IterateAccounts(consumer func(acm.Account) (stop bool)) (stopped bool, err error) {
	for _, acc := range tsr.Accounts {
		if consumer(acc) {
			return true, nil
		}
	}
	return false, nil
}

func (tsr *testState) IterateStorage(address acm.Address, consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {
	for key, value := range tsr.Storage[address] {
		if consumer(key, value) {
			return true, nil
		}
	}
	return false, nil
}

func (tsr *testState) GetAccount(address acm.Address) (acm.Account, error) {
	return tsr.Accounts[address], nil
}

func (tsr *testState) GetStorage(address acm.Address, key binary.Word256) (binary.Word256, error) {
	storage, ok := tsr.Storage[address]
	if !ok {
		return binary.Zero256, fmt.Errorf("could not find storage for account %s", address)
	}
	value, ok := storage[key]
	if !ok {
		return binary.Zero256, fmt.Errorf("could not find key %x for account %s", key, address)
	}
	return value, nil
}
