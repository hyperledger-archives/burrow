package state

import (
	"testing"

	"fmt"

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
	writeBackend := NewCache(NewMemoryState())
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

func TestStateCache_Miss(t *testing.T) {
	readBackend := testAccounts()
	cache := NewCache(readBackend)

	acc1Address := addressOf("acc1")
	acc1, err := cache.GetAccount(acc1Address)
	require.NoError(t, err)
	fmt.Println(acc1)

	acc1Exp := readBackend.Accounts[acc1Address]
	assert.Equal(t, acc1Exp, acc1)
	acc8, err := cache.GetAccount(addressOf("acc8"))
	require.NoError(t, err)
	assert.Nil(t, acc8)
}

func TestStateCache_UpdateAccount(t *testing.T) {
	// Build backend states for read and write
	backend := NewCache(NewMemoryState())
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
	// Build backend states for read and write
	readBackend := testAccounts()
	cache := NewCache(readBackend)

	acc := readBackend.Accounts[addressOf("acc1")]
	err := cache.RemoveAccount(acc.Address())
	require.NoError(t, err)

	dead, err := cache.GetAccount(acc.Address())
	assert.Nil(t, dead, err)
}

func TestStateCache_GetStorage(t *testing.T) {
}

func TestStateCache_SetStorage(t *testing.T) {
}

func TestStateCache_Sync(t *testing.T) {
}

func TestStateCache_get(t *testing.T) {
}

func testAccounts() *MemoryState {
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

func account(acc acm.Account, keyvals ...string) *MemoryState {
	ts := NewMemoryState()
	ts.Accounts[acc.Address()] = acc
	ts.Storage[acc.Address()] = make(map[binary.Word256]binary.Word256)
	for i := 0; i < len(keyvals); i += 2 {
		ts.Storage[acc.Address()][word(keyvals[i])] = word(keyvals[i+1])
	}
	return ts
}

func combine(states ...*MemoryState) *MemoryState {
	ts := NewMemoryState()
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
