package acmstate

import (
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/crypto/sha3"
	"github.com/hyperledger/burrow/permission"
	"github.com/tmthrgd/go-hex"
)

// AbiHash is the keccak hash for the ABI. This is to make the ABI content-addressed
type AbiHash [32]byte

func (h *AbiHash) Bytes() []byte {
	b := make([]byte, 32)
	copy(b, h[:])
	return b
}

func (ch *AbiHash) UnmarshalText(hexBytes []byte) error {
	bs, err := hex.DecodeString(string(hexBytes))
	if err != nil {
		return err
	}
	copy(ch[:], bs)
	return nil
}

func (ch AbiHash) MarshalText() ([]byte, error) {
	return []byte(ch.String()), nil
}

func (ch AbiHash) String() string {
	return hex.EncodeUpperToString(ch[:])
}

func GetAbiHash(abi string) (abihash AbiHash) {
	hash := sha3.NewKeccak256()
	hash.Write([]byte(abi))
	copy(abihash[:], hash.Sum(nil))
	return
}

// CodeHash is the keccak hash for the code for an account. This is used for the EVM CODEHASH opcode, and to find the
// correct ABI for a contract
type CodeHash [32]byte

func (h *CodeHash) Bytes() []byte {
	b := make([]byte, 32)
	copy(b, h[:])
	return b
}

func (ch *CodeHash) UnmarshalText(hexBytes []byte) error {
	bs, err := hex.DecodeString(string(hexBytes))
	if err != nil {
		return err
	}
	copy(ch[:], bs)
	return nil
}

func (ch CodeHash) MarshalText() ([]byte, error) {
	return []byte(ch.String()), nil
}

func (ch CodeHash) String() string {
	return hex.EncodeUpperToString(ch[:])
}

type AccountGetter interface {
	// Get an account by its address return nil if it does not exist (which should not be an error)
	GetAccount(address crypto.Address) (*acm.Account, error)
}

type AccountIterable interface {
	// Iterates through accounts calling passed function once per account, if the consumer
	// returns true the iteration breaks and returns true to indicate it iteration
	// was escaped
	IterateAccounts(consumer func(*acm.Account) error) (err error)
}

type AccountUpdater interface {
	// Updates the fields of updatedAccount by address, creating the account
	// if it does not exist
	UpdateAccount(updatedAccount *acm.Account) error
	// Remove the account at address
	RemoveAccount(address crypto.Address) error
}

type StorageGetter interface {
	// Retrieve a 32-byte value stored at key for the account at address, return Zero256 if key does not exist but
	// error if address does not
	GetStorage(address crypto.Address, key binary.Word256) (value []byte, err error)
}

type StorageSetter interface {
	// Store a 32-byte value at key for the account at address, setting to Zero256 removes the key
	SetStorage(address crypto.Address, key binary.Word256, value []byte) error
}

type StorageIterable interface {
	// Iterates through the storage of account ad address calling the passed function once per account,
	// if the iterator function returns true the iteration breaks and returns true to indicate it iteration
	// was escaped
	IterateStorage(address crypto.Address, consumer func(key binary.Word256, value []byte) error) (err error)
}

type AbiGetter interface {
	// Get an ABI by its hash. This is content-addressed
	GetAbi(abihash AbiHash) (string, error)
}

type AbiSetter interface {
	// Set an ABI according to it keccak-256 hash.
	SetAbi(abihash AbiHash, abi string) error
}

type AccountStats struct {
	AccountsWithCode    uint64
	AccountsWithoutCode uint64
}

type AccountStatsGetter interface {
	GetAccountStats() AccountStats
}

// Compositions

// Read-only account and storage state
type Reader interface {
	AccountGetter
	StorageGetter
	AbiGetter
}

type Iterable interface {
	AccountIterable
	StorageIterable
}

// Read and list account and storage state
type IterableReader interface {
	Iterable
	Reader
}

type IterableStatsReader interface {
	Iterable
	Reader
	AccountStatsGetter
}

type Writer interface {
	AccountUpdater
	StorageSetter
	AbiSetter
}

// Read and write account and storage state
type ReaderWriter interface {
	Reader
	Writer
}

type IterableReaderWriter interface {
	Iterable
	Reader
	Writer
}

func GlobalPermissionsAccount(getter AccountGetter) *acm.Account {
	acc, err := getter.GetAccount(acm.GlobalPermissionsAddress)
	if err != nil {
		panic("Could not get global permission account, but this must exist")
	}
	return acc
}

// Get global permissions from the account at GlobalPermissionsAddress
func GlobalAccountPermissions(getter AccountGetter) permission.AccountPermissions {
	if getter == nil {
		return permission.AccountPermissions{
			Roles: []string{},
		}
	}
	return GlobalPermissionsAccount(getter).Permissions
}
