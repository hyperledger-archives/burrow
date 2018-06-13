package state

import (
	"testing"

	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
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

	//Create account
	acc := readBackend.Accounts[addressOf("acc1")]

	//Get account from cache
	accOut, err := cache.GetAccount(acc.Address())
	require.NoError(t, err)
	cache.UpdateAccount(accOut)

	//Check that cache account matches original account
	assert.Equal(t, acc, *accOut)

	//Sync account to backend
	err = cache.Sync(writeBackend)
	require.NoError(t, err)

	//Get account from backend
	accOut, err = writeBackend.GetAccount(acc.Address())
	require.NotNil(t, accOut)
	assert.NoError(t, err)

	//Check that backend account matches original account
	assert.Equal(t, acc, *accOut)
}

func TestStateCache_Miss(t *testing.T) {
	readBackend := testAccounts()
	cache := NewCache(readBackend)

	acc1Address := addressOf("acc1")
	acc1, err := cache.GetAccount(acc1Address)
	require.NoError(t, err)
	fmt.Println(acc1)

	acc1Exp := readBackend.Accounts[acc1Address]
	assert.Equal(t, acc1Exp, *acc1)
	acc8, err := cache.GetAccount(addressOf("acc8"))
	require.NoError(t, err)
	assert.Nil(t, acc8)
}

func TestStateCache_UpdateAccount(t *testing.T) {
	// Build backend states for read and write
	backend := NewCache(NewMemoryState())
	cache := NewCache(backend)
	// Create acccount
	balance := uint64(24)
	accNew := acm.NewAccountFromSecret("accNew", permission.DefaultAccountPermissions)
	accNew.AddToBalance(balance)

	err := cache.UpdateAccount(accNew)
	require.NoError(t, err)

	// Check cache
	accNewOut, err := cache.GetAccount(accNew.Address())
	require.NoError(t, err)
	assert.Equal(t, balance, accNewOut.Balance())

	// Check not stored in backend
	accNewOut, err = backend.GetAccount(accNew.Address())
	require.NoError(t, err)
	assert.Nil(t, accNewOut)

	// Check syncs to backend
	err = cache.Sync(backend)
	require.NoError(t, err)
	accNewOut, err = backend.GetAccount(accNew.Address())
	require.NoError(t, err)
	assert.Equal(t, balance, accNewOut.Balance())

	// Alter in cache
	addBalance := uint64(100029)
	accNew.AddToBalance(addBalance)
	err = cache.UpdateAccount(accNew)
	require.NoError(t, err)

	// Check cache
	accNewOut, err = cache.GetAccount(accNew.Address())
	require.NoError(t, err)
	assert.Equal(t, balance+addBalance, accNewOut.Balance())

	// Check backend unchanged
	accNewOut, err = backend.GetAccount(accNew.Address())
	require.NoError(t, err)
	assert.Equal(t, balance, accNewOut.Balance())
}

func TestStateCache_RemoveAccount(t *testing.T) {
	// Build backend states for read and write
	backend := NewCache(NewMemoryState())
	cache := NewCache(backend)

	//Create new account
	newAcc := acm.NewAccountFromSecret("newAcc", permission.DefaultAccountPermissions)
	newAcc.AddToBalance(100)
	err := cache.UpdateAccount(newAcc)
	require.NoError(t, err)

	//Sync account to backend
	err = cache.Sync(backend)
	require.NoError(t, err)

	//Check for account in cache
	newAccOut, err := cache.GetAccount(newAcc.Address())
	require.NoError(t, err)
	require.NotNil(t, newAccOut)

	//Check for account in backend
	newAccOut, err = backend.GetAccount(newAcc.Address())
	require.NoError(t, err)
	require.NotNil(t, newAccOut)

	//Remove account from cache and backend
	err = cache.RemoveAccount(newAcc.Address())
	require.NoError(t, err)
	err = cache.Sync(backend)
	require.NoError(t, err)

	//Check that account is removed from cache
	newAccOut, err = cache.GetAccount(newAcc.Address())
	require.NoError(t, err)
	require.Nil(t, newAccOut)

	//Check that account is removed from backend
	newAccOut, err = backend.GetAccount(newAcc.Address())
	require.NoError(t, err)
	require.Nil(t, newAccOut)
}

func TestStateCache_GetStorage(t *testing.T) {
	// Build backend states for read and write
	readBackend := testAccounts()
	writeBackend := NewCache(NewMemoryState())
	cache := NewCache(readBackend)

	//Create account
	acc := readBackend.Accounts[addressOf("acc1")]

	//Get storage from cache
	accStorage, err := cache.GetStorage(acc.Address(), word("I AM A KEY"))
	require.NoError(t, err)
	cache.UpdateAccount(&acc)

	//Check for correct cache storage value
	assert.Equal(t, word("NO YOU ARE A KEY"), accStorage)

	//Sync account to backend
	err = cache.Sync(writeBackend)
	require.NoError(t, err)

	//Get storage from backend
	accStorage, err = writeBackend.GetStorage(acc.Address(), word("I AM A KEY"))
	assert.NoError(t, err)
	require.NotNil(t, accStorage)

	//Check for correct backend storage value
	assert.Equal(t, word("NO YOU ARE A KEY"), accStorage)
}

func TestStateCache_SetStorage(t *testing.T) {
	// Build backend states for read and write
	backend := NewCache(NewMemoryState())
	cache := NewCache(backend)

	//Create new account and set its storage in cache
	newAcc := acm.NewAccountFromSecret("newAcc", permission.DefaultAccountPermissions)
	newAcc.AddToBalance(100)
	err := cache.SetStorage(newAcc.Address(), word("What?"), word("Huh?"))
	require.Error(t, err)
	err = cache.UpdateAccount(newAcc)
	require.NoError(t, err)
	err = cache.SetStorage(newAcc.Address(), word("What?"), word("Huh?"))
	require.NoError(t, err)

	//Check for correct cache storage value
	newAccStorage, err := cache.GetStorage(newAcc.Address(), word("What?"))
	require.NoError(t, err)
	assert.Equal(t, word("Huh?"), newAccStorage)

	//Sync account to backend
	err = cache.Sync(backend)
	require.NoError(t, err)

	//Check for correct backend storage value
	newAccStorage, err = backend.GetStorage(newAcc.Address(), word("What?"))
	require.NoError(t, err)
	assert.Equal(t, word("Huh?"), newAccStorage)
}

func TestStateCache_Sync(t *testing.T) {
	// Build backend states for read and write
	backend := NewCache(NewMemoryState())
	cache := NewCache(backend)

	//Create new account
	balance := uint64(24)
	newAcc := acm.NewAccountFromSecret("newAcc", permission.DefaultAccountPermissions)
	newAcc.AddToBalance(balance)

	//Update cache with account changes
	err := cache.UpdateAccount(newAcc)
	require.NoError(t, err)

	//Set storage for account
	err = cache.SetStorage(newAcc.Address(), word("God save"), word("the queen!"))
	require.NoError(t, err)

	//Confirm changes to account balance in cache
	newAccOut, err := cache.GetAccount(newAcc.Address())
	require.NoError(t, err)
	assert.Equal(t, balance, newAccOut.Balance())

	//Confirm changes to account storage in cache
	newAccStorage, err := cache.GetStorage(newAcc.Address(), word("God save"))
	require.NoError(t, err)
	assert.Equal(t, word("the queen!"), newAccStorage)

	//Sync account to backend
	err = cache.Sync(backend)
	require.NoError(t, err)

	//Confirm changes to account balance synced correctly to backend
	newAccOut, err = backend.GetAccount(newAcc.Address())
	require.NoError(t, err)
	assert.Equal(t, balance, newAccOut.Balance())

	//Confirm changes to account storage synced correctly to backend
	newAccStorage, err = cache.GetStorage(newAcc.Address(), word("God save"))
	require.NoError(t, err)
	assert.Equal(t, word("the queen!"), newAccStorage)

	//Remove account from cache
	err = cache.RemoveAccount(newAcc.Address())
	require.NoError(t, err)

	//Sync account removal to backend
	err = cache.Sync(backend)
	require.NoError(t, err)

	//Check that account removal synced correctly to backend
	newAccOut, err = backend.GetAccount(newAcc.Address())
	require.NoError(t, err)
	require.Nil(t, newAccOut)
}

func TestStateCache_get(t *testing.T) {
	backend := NewCache(NewMemoryState())
	cache := NewCache(backend)

	//Create new account
	newAcc := acm.NewAccountFromSecret("newAcc", permission.DefaultAccountPermissions)
	newAcc.AddToBalance(100)

	//Add new account to cache
	err := cache.UpdateAccount(newAcc)
	require.NoError(t, err)

	//Check that get returns an account from cache
	newAccOut, err := cache.GetAccount(newAcc.Address())
	require.NoError(t, err)
	require.NotNil(t, newAccOut)

	//Sync account to backend
	err = cache.Sync(backend)
	require.NoError(t, err)

	//Check that get returns an account from backend
	newAccOut, err = backend.GetAccount(newAcc.Address())
	require.NoError(t, err)
	require.NotNil(t, newAccOut)

}

func testAccounts() *MemoryState {
	code, _ := acm.NewBytecode(asm.PUSH1, 0x20)

	acc1 := acm.NewAccountFromSecret("acc1", permission.AllAccountPermissions)
	acc2 := acm.NewContractAccountFromSecret("acc2", permission.DefaultAccountPermissions)
	acc1.AddToBalance(100)
	acc2.AddToBalance(100)
	acc2.SetCode(code)

	cache := combine(
		account(acc1, "I AM A KEY", "NO YOU ARE A KEY"),
		account(acc2, "ducks", "have lucks",
			"chickens", "just cluck"),
	)
	return cache
}

func addressOf(secret string) crypto.Address {
	return acm.NewAccountFromSecret(secret, permission.DefaultAccountPermissions).Address()
}

func account(acc *acm.Account, keyvals ...string) *MemoryState {
	ts := NewMemoryState()
	ts.Accounts[acc.Address()] = *acc
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
