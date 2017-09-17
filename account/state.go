package account

import (
	"github.com/hyperledger/burrow/word"
)

type Creator interface {
	// Create an account as a child of the creatorAccount deriving the new
	// accounts address from the creator's address and updating the creator's
	// sequence number
	CreateAccount(creatorAccount *ConcreteAccount) *ConcreteAccount
}

type Getter interface {
	// Get an account by its address
	GetAccount(address Address) *ConcreteAccount
}

type Updater interface {
	Getter
	// Updates the fields of updatedAccount by address, creating the account
	// if it does not exist
	UpdateAccount(updatedAccount *ConcreteAccount)
	// Remove the account at address
	RemoveAccount(address Address)
}

type StorageGetter interface {
	// Retrieve a 32-byte value stored at key for the account at address
	GetStorage(address Address, key word.Word256) (value word.Word256)
}

type Storage interface {
	StorageGetter
	// Store a 32-byte value at key for the account at address
	SetStorage(address Address, key word.Word256, value word.Word256)
}

// Read-write account and storage state
type UpdaterAndStorage interface {
	Updater
	Storage
}

// Read-only account and storage state
type GetterAndStorageGetter interface {
	Getter
	StorageGetter
}
