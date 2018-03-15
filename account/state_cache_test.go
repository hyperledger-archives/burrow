package account

import (
	"testing"

	"fmt"

	"github.com/hyperledger/burrow/binary"
)

type testStateReader struct {
	Accounts map[Address]Account
	Storage  map[Address]map[binary.Word256]binary.Word256
}

func accountAndStorage(account Account, keyvals ...binary.Word256) *testStateReader {
	return &testStateReader{}

}

func (tsr *testStateReader) GetAccount(address Address) (Account, error) {
	account, ok := tsr.Accounts[address]
	if !ok {
		return nil, fmt.Errorf("could not find account %s", address)
	}
	return account, nil
}

func (tsr *testStateReader) GetStorage(address Address, key binary.Word256) (binary.Word256, error) {
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

var _ StateReader = &testStateReader{}

func TestStateCache_GetAccount(t *testing.T) {
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
