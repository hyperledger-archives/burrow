package acmstate

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
)

type MemoryState struct {
	Accounts  map[crypto.Address]*acm.Account
	Storage   map[crypto.Address]map[binary.Word256][]byte
	Metadatas map[MetadataHash]string
}

var _ IterableReaderWriter = &MemoryState{}

// Get an in-memory state IterableReader
func NewMemoryState() *MemoryState {
	return &MemoryState{
		Accounts:  make(map[crypto.Address]*acm.Account),
		Storage:   make(map[crypto.Address]map[binary.Word256][]byte),
		Metadatas: make(map[MetadataHash]string),
	}
}

func (ms *MemoryState) GetAccount(address crypto.Address) (*acm.Account, error) {
	return ms.Accounts[address], nil
}

func (ms *MemoryState) UpdateAccount(updatedAccount *acm.Account) error {
	if updatedAccount == nil {
		return fmt.Errorf("UpdateAccount passed nil account in MemoryState")
	}
	ms.Accounts[updatedAccount.GetAddress()] = updatedAccount
	return nil
}

func (ms *MemoryState) GetMetadata(metahash MetadataHash) (string, error) {
	return ms.Metadatas[metahash], nil
}

func (ms *MemoryState) SetMetadata(metahash MetadataHash, metadata string) error {
	ms.Metadatas[metahash] = metadata
	return nil
}

func (ms *MemoryState) RemoveAccount(address crypto.Address) error {
	delete(ms.Accounts, address)
	return nil
}

func (ms *MemoryState) GetStorage(address crypto.Address, key binary.Word256) ([]byte, error) {
	storage, ok := ms.Storage[address]
	if !ok {
		return []byte{}, fmt.Errorf("could not find storage for account %s", address)
	}
	value, ok := storage[key]
	if !ok {
		return []byte{}, fmt.Errorf("could not find key %x for account %s", key, address)
	}
	return value, nil
}

func (ms *MemoryState) SetStorage(address crypto.Address, key binary.Word256, value []byte) error {
	storage, ok := ms.Storage[address]
	if !ok {
		storage = make(map[binary.Word256][]byte)
		ms.Storage[address] = storage
	}
	storage[key] = value
	return nil
}

func (ms *MemoryState) IterateAccounts(consumer func(*acm.Account) error) (err error) {
	for _, acc := range ms.Accounts {
		if err := consumer(acc); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MemoryState) IterateStorage(address crypto.Address, consumer func(key binary.Word256, value []byte) error) (err error) {
	for key, value := range ms.Storage[address] {
		if err := consumer(key, value); err != nil {
			return err
		}
	}
	return nil
}
