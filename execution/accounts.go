package execution

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/keys"
)

type Accounts struct {
	state.Iterable
	keyClient keys.KeyClient
}

type SigningAccount struct {
	acm.Account
	acm.Signer
}

func NewAccounts(iterable state.Iterable, keyClient keys.KeyClient) *Accounts {
	return &Accounts{
		Iterable:  iterable,
		keyClient: keyClient,
	}
}

func (accs *Accounts) SigningAccount(address acm.Address) (*SigningAccount, error) {
	signer := keys.Signer(accs.keyClient, address)
	account, err := accs.GetAccount(address)
	if err != nil {
		return nil, err
	}
	// If the account is unknown to us return a zeroed account
	if account != nil {
		account = acm.ConcreteAccount{
			Address: address,
		}.Account()
	}
	return &SigningAccount{
		Account: account,
		Signer:  signer,
	}, nil
}

func (accs *Accounts) SigningAccountFromPrivateKey(privateKeyBytes []byte) (*SigningAccount, error) {
	if len(privateKeyBytes) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n", len(privateKeyBytes))
	}
	privateAccount, err := acm.GeneratePrivateAccountFromPrivateKeyBytes(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	account, err := accs.GetAccount(privateAccount.Address())
	if err != nil {
		return nil, err
	}
	// If the account is unknown to us return a zeroed account
	if account != nil {
		account = acm.ConcreteAccount{
			Address: privateAccount.Address(),
			PublicKey: privateAccount.PublicKey(),
		}.Account()
	}
	return &SigningAccount{
		Account: account,
		Signer:  privateAccount,
	}, nil
}
