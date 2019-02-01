package evm

import (
	"testing"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/stretchr/testify/assert"
)

func TestState_PushError(t *testing.T) {
	st := NewState(newAppState(), newBlockchainInfo())
	// This will be a wrapped nil - it should not register as first error
	var ex errors.CodedError = (*errors.Exception)(nil)
	st.PushError(ex)
	// This one should
	realErr := errors.ErrorCodef(errors.ErrorCodeInsufficientBalance, "real error")
	st.PushError(realErr)
	assert.True(t, realErr.Equal(st.Error()))
}

func TestState_CreateAccount(t *testing.T) {
	st := NewState(newAppState(), newBlockchainInfo())
	address := newAddress("frogs")
	st.CreateAccount(address)
	require.Nil(t, st.Error())
	st.CreateAccount(address)
	assertErrorCode(t, errors.ErrorCodeDuplicateAddress, st.Error())

	st = NewState(newAppState(), newBlockchainInfo())
	st.CreateAccount(address)
	require.Nil(t, st.Error())
	st.InitCode(address, []byte{1, 2, 3})
	require.Nil(t, st.Error())
}

func TestState_Sync(t *testing.T) {
	backend := acmstate.NewCache(newAppState())
	st := NewState(backend, newBlockchainInfo())
	address := newAddress("frogs")

	st.CreateAccount(address)
	amt := uint64(1232)
	st.AddToBalance(address, amt)

	var err error
	err = st.Sync()
	require.Nil(t, err)
	acc, err := backend.GetAccount(address)
	require.NoError(t, err)
	assert.Equal(t, acc.Balance, amt)
}

func TestState_NewCache(t *testing.T) {
	st := NewState(newAppState(), newBlockchainInfo())
	address := newAddress("frogs")

	cache := st.NewCache()
	cache.CreateAccount(address)
	amt := uint64(1232)
	cache.AddToBalance(address, amt)

	var err error
	assert.Equal(t, uint64(0), st.GetBalance(address))
	require.Nil(t, st.Error())

	// Sync through to cache
	err = cache.Sync()
	require.NoError(t, err)
	assert.Equal(t, amt, st.GetBalance(address))
	require.Nil(t, st.Error())

	cache = st.NewCache(acmstate.ReadOnly).NewCache()
	require.Nil(t, st.Error())
	cache.AddToBalance(address, amt)
	assertErrorCode(t, errors.ErrorCodeIllegalWrite, cache.Error())
}

func newBlockchainInfo() *bcm.Blockchain {
	return &bcm.Blockchain{}
}
