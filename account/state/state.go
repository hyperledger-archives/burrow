package state

import (
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	ptypes "github.com/hyperledger/burrow/permission/types"
)

type AccountGetter interface {
	// Get an account by its address return nil if it does not exist (which should not be an error)
	GetAccount(address acm.Address) (acm.Account, error)
}

type AccountIterable interface {
	// Iterates through accounts calling passed function once per account, if the consumer
	// returns true the iteration breaks and returns true to indicate it iteration
	// was escaped
	IterateAccounts(consumer func(acm.Account) (stop bool)) (stopped bool, err error)
}

type AccountUpdater interface {
	// Updates the fields of updatedAccount by address, creating the account
	// if it does not exist
	UpdateAccount(updatedAccount acm.Account) error
	// Remove the account at address
	RemoveAccount(address acm.Address) error
}

type StorageGetter interface {
	// Retrieve a 32-byte value stored at key for the account at address, return Zero256 if key does not exist but
	// error if address does not
	GetStorage(address acm.Address, key binary.Word256) (value binary.Word256, err error)
}

type StorageSetter interface {
	// Store a 32-byte value at key for the account at address, setting to Zero256 removes the key
	SetStorage(address acm.Address, key, value binary.Word256) error
}

type StorageIterable interface {
	// Iterates through the storage of account ad address calling the passed function once per account,
	// if the iterator function returns true the iteration breaks and returns true to indicate it iteration
	// was escaped
	IterateStorage(address acm.Address, consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error)
}

// Compositions

// Read-only account and storage state
type Reader interface {
	AccountGetter
	StorageGetter
}

// Read and list account and storage state
type Iterable interface {
	Reader
	AccountIterable
	StorageIterable
}

// Read and write account and storage state
type Writer interface {
	Reader
	AccountUpdater
	StorageSetter
}

type IterableWriter interface {
	Reader
	AccountUpdater
	StorageSetter
	AccountIterable
	StorageIterable
}

func GetMutableAccount(getter AccountGetter, address acm.Address) (acm.MutableAccount, error) {
	acc, err := getter.GetAccount(address)
	if err != nil {
		return nil, err
	}
	return acm.AsMutableAccount(acc), nil
}

func GlobalPermissionsAccount(getter AccountGetter) acm.Account {
	acc, err := getter.GetAccount(acm.GlobalPermissionsAddress)
	if err != nil {
		panic("Could not get global permission account, but this must exist")
	}
	return acc
}

// Get global permissions from the account at GlobalPermissionsAddress
func GlobalAccountPermissions(getter AccountGetter) ptypes.AccountPermissions {
	if getter == nil {
		return ptypes.AccountPermissions{
			Roles: []string{},
		}
	}
	return GlobalPermissionsAccount(getter).Permissions()
}
