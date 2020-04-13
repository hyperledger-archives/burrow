package native

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmthrgd/go-bitset"
)

func TestState_CreateAccount(t *testing.T) {
	st := acmstate.NewMemoryState()
	address := AddressFromName("frogs")
	err := CreateAccount(st, address)
	require.NoError(t, err)
	err = CreateAccount(st, address)
	require.Error(t, err)
	require.Equal(t, errors.Codes.DuplicateAddress, errors.GetCode(err))

	st = acmstate.NewMemoryState()
	err = CreateAccount(st, address)
	require.NoError(t, err)
	err = InitEVMCode(st, address, []byte{1, 2, 3})
	require.NoError(t, err)
}

func TestState_Sync(t *testing.T) {
	backend := acmstate.NewCache(acmstate.NewMemoryState())
	st := engine.NewCallFrame(backend)
	address := AddressFromName("frogs")

	err := CreateAccount(st, address)
	require.NoError(t, err)
	amt := uint64(1232)
	addToBalance(t, st, address, amt)
	err = st.Sync()
	require.NoError(t, err)

	acc, err := backend.GetAccount(address)
	require.NoError(t, err)
	assert.Equal(t, acc.Balance, amt)
}

func TestState_NewCache(t *testing.T) {
	st := engine.NewCallFrame(acmstate.NewMemoryState())
	address := AddressFromName("frogs")

	cache, err := st.NewFrame()
	require.NoError(t, err)
	err = CreateAccount(cache, address)
	require.NoError(t, err)
	amt := uint64(1232)
	addToBalance(t, cache, address, amt)

	acc, err := st.GetAccount(address)
	require.NoError(t, err)
	require.Nil(t, acc)

	// Sync through to cache
	err = cache.Sync()
	require.NoError(t, err)

	acc, err = mustAccount(cache, address)
	require.NoError(t, err)
	assert.Equal(t, amt, acc.Balance)

	cache, err = cache.NewFrame(acmstate.ReadOnly)
	require.NoError(t, err)
	cache, err = cache.NewFrame()
	require.NoError(t, err)
	err = UpdateAccount(cache, address, func(account *acm.Account) error {
		return account.AddToBalance(amt)
	})
	require.Error(t, err)
	require.Equal(t, errors.Codes.IllegalWrite, errors.GetCode(err))
}

func TestOpcodeBitset(t *testing.T) {
	tests := []struct {
		name   string
		code   []byte
		bitset bitset.Bitset
	}{
		{
			name:   "Only one real JUMPDEST",
			code:   bc.MustSplice(asm.PUSH2, 1, asm.JUMPDEST, asm.ADD, asm.JUMPDEST),
			bitset: mkBitset("10011"),
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := opcodeBitset(tt.code); !reflect.DeepEqual(got, tt.bitset) {
				t.Errorf("opcodeBitset() = %v, want %v", got, tt.bitset)
			}
		})
	}
}

func mkBitset(binaryString string) bitset.Bitset {
	length := uint(len(binaryString))
	bs := bitset.New(length)
	for i := uint(0); i < length; i++ {
		switch binaryString[i] {
		case '1':
			bs.Set(i)
		case '0':
		case ' ':
			i++
		default:
			panic(fmt.Errorf("mkBitset() expects a string containing only 1s, 0s, and spaces"))
		}
	}
	return bs
}

func addToBalance(t testing.TB, st acmstate.ReaderWriter, address crypto.Address, amt uint64) {
	err := UpdateAccount(st, address, func(account *acm.Account) error {
		return account.AddToBalance(amt)
	})
	require.NoError(t, err)
}
