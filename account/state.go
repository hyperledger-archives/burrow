package account

import (
	"github.com/hyperledger/burrow/binary"
)

type Getter interface {
	// Get an account by its address return nil if it does not exist (which should not be an error)
	GetAccount(address Address) (Account, error)
}

type Iterable interface {
	// Iterates through accounts calling passed function once per account, if the consumer
	// returns true the iteration breaks and returns true to indicate it iteration
	// was escaped
	IterateAccounts(consumer func(Account) (stop bool)) (stopped bool, err error)
}

type Updater interface {
	// Updates the fields of updatedAccount by address, creating the account
	// if it does not exist
	UpdateAccount(updatedAccount Account) error
	// Remove the account at address
	RemoveAccount(address Address) error
}

type StorageGetter interface {
	// Retrieve a 32-byte value stored at key for the account at address, return Zero256 if key does not exist but
	// error if address does not
	GetStorage(address Address, key binary.Word256) (value binary.Word256, err error)
}

type StorageSetter interface {
	// Store a 32-byte value at key for the account at address
	SetStorage(address Address, key, value binary.Word256) error
}

type StorageIterable interface {
	// Iterates through the storage of account ad address calling the passed function once per account,
	// if the iterator function returns true the iteration breaks and returns true to indicate it iteration
	// was escaped
	IterateStorage(address Address, consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error)
}

// Compositions

// Read-only account and storage state
type StateReader interface {
	Getter
	StorageGetter
}

// Read and list account and storage state
type StateIterable interface {
	StateReader
	Iterable
	StorageIterable
}

// Read and write account and storage state
type StateWriter interface {
	StateReader
	Updater
	StorageSetter
}

type IterableStateWriter interface {
	StateReader
	Updater
	StorageSetter
	Iterable
	StorageIterable
}
