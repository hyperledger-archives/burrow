package genesis

import (
	"testing"

	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
)

func TestMakeGenesisDocFromAccounts(t *testing.T) {
	genDoc := MakeGenesisDocFromAccounts("test-chain", nil,
		accountMap("Tinkie-winkie", "Lala", "Po", "Dipsy"),
		validatorMap("Foo", "Bar", "Baz"),
	)

	// Check we have matching serialisation after a round trip
	bs, err := genDoc.JSONBytes()
	assert.NoError(t, err)

	genDocOut, err := GenesisDocFromJSON(bs)
	assert.NoError(t, err)

	bsOut, err := genDocOut.JSONBytes()
	assert.NoError(t, err)

	assert.Equal(t, bs, bsOut)
	assert.Equal(t, genDoc.Hash(), genDocOut.Hash())
	fmt.Println(string(bs))
}

func accountMap(names ...string) map[string]acm.Account {
	accounts := make(map[string]acm.Account, len(names))
	for _, name := range names {
		accounts[name] = accountFromName(name)
	}
	return accounts
}

func validatorMap(names ...string) map[string]acm.Validator {
	validators := make(map[string]acm.Validator, len(names))
	for _, name := range names {
		validators[name] = acm.AsValidator(accountFromName(name))
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
