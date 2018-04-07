package execution

import (
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/keys"
)

type Accounts struct {
	state.Reader
	keyClient keys.KeyClient
}

type SigningAccount struct {
	acm.Account
	acm.Signer
}

func NewAccounts(reader state.Reader, keyClient keys.KeyClient) *Accounts {
	return &Accounts{
		Reader:    reader,
		keyClient: keyClient,
	}
}

func (accs *Accounts) SigningAccount(address acm.Address) (*SigningAccount, error) {
	signer := keys.Signer(accs.keyClient, address)
	account, err := state.GetMutableAccount(accs.Reader, address)
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

func (accs *Accounts) SigningAccountFromPrivateKey(privateKeyBytes []byte) (*SigningAccount, error) {
	if len(privateKeyBytes) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n", len(privateKeyBytes))
	}
	privateAccount, err := acm.GeneratePrivateAccountFromPrivateKeyBytes(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	account, err := state.GetMutableAccount(accs, privateAccount.Address())
	if err != nil {
		return nil, err
	}
	// If the account is unknown to us return zeroed account for the address derived from the private key
	if account == nil {
		account = acm.ConcreteAccount{
			Address: privateAccount.Address(),
		}.MutableAccount()
	}
	// Set the public key in case it was not known previously (needed for signing with an unseen account)
	account.SetPublicKey(privateAccount.PublicKey())
	return &SigningAccount{
		Account: account,
		Signer:  privateAccount,
	}, nil
}
