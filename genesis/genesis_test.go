package genesis

import (
	"fmt"
	"testing"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
)

func TestMakeGenesisDocFromAccounts(t *testing.T) {
	var genesisTime = time.Now()

	genDoc := MakeGenesisDocFromAccounts("test-chain", nil, genesisTime,
		permission.DefaultAccountPermissions.Clone(),
		accountList("Tinkie-winkie", "Lala", "Po", "Dipsy"),
		validatorList("Foo", "Bar", "Baz"),
	)

	genDoc2 := MakeGenesisDocFromAccounts("test-chain", nil, genesisTime,
		permission.DefaultAccountPermissions.Clone(),
		accountList("Lala", "Po", "Tinkie-winkie", "Dipsy"),
		validatorList("Bar", "Baz", "Foo"),
	)

	genDoc4 := MakeGenesisDocFromAccounts("test-chain1", nil, genesisTime,
		permission.DefaultAccountPermissions.Clone(),
		accountList("Lala", "Po", "Tinkie-winkie", "Dipsy"),
		validatorList("Bar", "Baz", "Foo"),
	)

	// Check we have matching serialisation after a round trip
	bs, err := genDoc.JSONBytes()
	assert.NoError(t, err)

	genDoc3, err := GenesisDocFromJSON(bs)
	assert.NoError(t, err)

	bsOut, err := genDoc3.JSONBytes()
	assert.NoError(t, err)

	assert.Equal(t, bs, bsOut)
	assert.Equal(t, genDoc.Hash(), genDoc2.Hash())
	assert.Equal(t, genDoc.Hash(), genDoc3.Hash())
	assert.NotEqual(t, genDoc.Hash(), genDoc4.Hash())
	fmt.Println(string(bs))
}

func accountList(names ...string) []acm.Account {
	accounts := make([]acm.Account, len(names))
	for i, name := range names {
		accounts[i] = accountFromName(name)
	}
	return accounts
}

func validatorList(names ...string) []acm.Validator {
	validators := make([]acm.Validator, len(names))
	for i, name := range names {
		account := accountFromName(name)
		validators[i] = acm.NewValidator(account.PublicKey(), account.Balance(), 1)
	}
	return validators
}

func accountFromName(name string) acm.Account {
	ca := acm.NewConcreteAccountFromSecret(name)
	for _, c := range name {
		ca.Balance += uint64(c)
	}
	ca.Permissions = permission.AllAccountPermissions.Clone()
	return ca.Account()
}
