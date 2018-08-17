package execution

import (
	"sync"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys"
	burrow_sync "github.com/hyperledger/burrow/sync"
)

// Accounts pairs an underlying state.Reader with a KeyClient to provide a signing variant of an account
// it also maintains a lock over addresses to provide a linearisation of signing events using SequentialSigningAccount
type Accounts struct {
	burrow_sync.RingMutex
	state.Reader
	keyClient keys.KeyClient
}

type SigningAccount struct {
	acm.Account
	crypto.Signer
}

type SequentialSigningAccount struct {
	Address       crypto.Address
	accountLocker sync.Locker
	getter        func() (*SigningAccount, error)
}

func NewAccounts(reader state.Reader, keyClient keys.KeyClient, mutexCount int) *Accounts {
	return &Accounts{
		RingMutex: *burrow_sync.NewRingMutexNoHash(mutexCount),
		Reader:    reader,
		keyClient: keyClient,
	}
}
func (accs *Accounts) SigningAccount(address crypto.Address) (*SigningAccount, error) {
	signer, err := keys.AddressableSigner(accs.keyClient, address)
	if err != nil {
		return nil, err
	}
	account, err := state.GetMutableAccount(accs, address)
	if err != nil {
		return nil, err
	}
	// If the account is unknown to us return a zeroed account
	if account == nil {
		account = acm.ConcreteAccount{
			Address: address,
		}.MutableAccount()
	}
	pubKey, err := accs.keyClient.PublicKey(address)
	if err != nil {
		return nil, err
	}
	account.SetPublicKey(pubKey)
	return &SigningAccount{
		Account: account,
		Signer:  signer,
	}, nil
}

func (accs *Accounts) SequentialSigningAccount(address crypto.Address) (*SequentialSigningAccount, error) {
	return &SequentialSigningAccount{
		Address:       address,
		accountLocker: accs.Mutex(address.Bytes()),
		getter: func() (*SigningAccount, error) {
			return accs.SigningAccount(address)
		},
	}, nil
}

type UnlockFunc func()

func (ssa *SequentialSigningAccount) Lock() (*SigningAccount, UnlockFunc, error) {
	ssa.accountLocker.Lock()
	account, err := ssa.getter()
	if err != nil {
		ssa.accountLocker.Unlock()
		return nil, nil, err
	}
	return account, ssa.accountLocker.Unlock, err
}
