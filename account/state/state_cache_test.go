package state

import (
	"fmt"
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateCache_GetAccount(t *testing.T) {
	// Build backend states for read and write
	readBackend := testAccounts()
	writeBackend := NewCache(newTestState())
	cache := NewCache(readBackend)

	acc := readBackend.Accounts[addressOf("acc1")]
	accOut, err := cache.GetAccount(acc.Address())
	require.NoError(t, err)
	cache.UpdateAccount(accOut)
	assert.Equal(t, acm.AsConcreteAccount(acc), acm.AsConcreteAccount(accOut))

	err = cache.Sync(writeBackend)
	require.NoError(t, err)
	accOut, err = writeBackend.GetAccount(acc.Address())
	require.NotNil(t, accOut)
	assert.NoError(t, err)
	assert.Equal(t, acm.AsConcreteAccount(acc), acm.AsConcreteAccount(accOut))
}

func TestStateCache_UpdateAccount(t *testing.T) {
	// Build backend states for read and write
	backend := NewCache(newTestState())
	cache := NewCache(backend)
	// Create acccount
	accNew := acm.NewConcreteAccountFromSecret("accNew")
	balance := uint64(24)
	accNew.Balance = balance
	err := cache.UpdateAccount(accNew.Account())
	require.NoError(t, err)

	// Check cache
	accNewOut, err := cache.GetAccount(accNew.Address)
	require.NoError(t, err)
	assert.Equal(t, balance, accNewOut.Balance())

	// Check not stored in backend
	accNewOut, err = backend.GetAccount(accNew.Address)
	require.NoError(t, err)
	assert.Nil(t, accNewOut)

	// Check syncs to backend
	err = cache.Sync(backend)
	require.NoError(t, err)
	accNewOut, err = backend.GetAccount(accNew.Address)
	require.NoError(t, err)
	assert.Equal(t, balance, accNewOut.Balance())

	// Alter in cache
	newBalance := uint64(100029)
	accNew.Balance = newBalance
	err = cache.UpdateAccount(accNew.Account())
	require.NoError(t, err)

	// Check cache
	accNewOut, err = cache.GetAccount(accNew.Address)
	require.NoError(t, err)
	assert.Equal(t, newBalance, accNewOut.Balance())

	// Check backend unchanged
	accNewOut, err = backend.GetAccount(accNew.Address)
	require.NoError(t, err)
	assert.Equal(t, balance, accNewOut.Balance())
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

func testAccounts() *testState {
	acc1 := acm.NewConcreteAccountFromSecret("acc1")
	acc1.Permissions.Base.Perms = permission.AddRole | permission.Send
	acc1.Permissions.Base.SetBit = acc1.Permissions.Base.Perms

	acc2 := acm.NewConcreteAccountFromSecret("acc2")
	acc2.Permissions.Base.Perms = permission.AddRole | permission.Send
	acc2.Permissions.Base.SetBit = acc1.Permissions.Base.Perms
	acc2.Code, _ = acm.NewBytecode(asm.PUSH1, 0x20)

	cache := combine(
		account(acc1.Account(), "I AM A KEY", "NO YOU ARE A KEY"),
		account(acc2.Account(), "ducks", "have lucks",
			"chickens", "just cluck"),
	)
	return cache
}

func addressOf(secret string) acm.Address {
	return acm.NewConcreteAccountFromSecret(secret).Address
}

// testState StateIterable

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
