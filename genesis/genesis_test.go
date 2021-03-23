package genesis

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
)

var genesisTime, _ = time.Parse("02-01-2006", "27-10-2017")

func TestMakeGenesisDocFromAccounts(t *testing.T) {
	genDoc := MakeGenesisDocFromAccounts("test-chain", nil, genesisTime,
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
func TestGenesisStability(t *testing.T) {
	genDoc := MakeGenesisDocFromAccounts("test-chain", nil, genesisTime,
		accountMap("Tinkie-winkie", "Lala", "Po", "Dipsy"),
		validatorMap("Foo", "Bar", "Baz"),
	)

	require.Equal(t, expectedGenesisJSON, genDoc.JSONString())

	require.Equal(t, "C5B64E6AD231221C328271ADCE401AA11F9DF12830F7DA2FC3B2C923E929C532", genDoc.Hash().String())
}

func accountMap(names ...string) map[string]*acm.Account {
	accounts := make(map[string]*acm.Account, len(names))
	for _, name := range names {
		accounts[name] = accountFromName(name)
	}
	return accounts
}

func validatorMap(names ...string) map[string]*validator.Validator {
	validators := make(map[string]*validator.Validator, len(names))
	for _, name := range names {
		acc := accountFromName(name)
		validators[name] = validator.FromAccount(acc, acc.Balance)
	}
	return validators
}

func accountFromName(name string) *acm.Account {
	ca := acm.NewAccountFromSecret(name)
	for _, c := range name {
		ca.Balance += uint64(c)
	}
	ca.Permissions = permission.AllAccountPermissions.Clone()
	return ca
}

// For genesis stability test
const expectedGenesisJSON = `{
	"GenesisTime": "2017-10-27T00:00:00Z",
	"ChainName": "test-chain",
	"Params": {
		"ProposalThreshold": 0
	},
	"GlobalPermissions": {
		"Base": {
			"Perms": "send | call | createContract | createAccount | bond | name | proposal | input | batch | hasBase | hasRole",
			"SetBit": "root | send | call | createContract | createAccount | bond | name | proposal | input | batch | identify | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole"
		}
	},
	"Accounts": [
		{
			"Address": "410427F4A361B958C97B47D81DAFDCF0A2B6503D",
			"PublicKey": null,
			"Amount": 521,
			"Name": "Dipsy",
			"Permissions": {
				"Base": {
					"Perms": "root | send | call | createContract | createAccount | bond | name | proposal | input | batch | identify | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole",
					"SetBit": "root | send | call | createContract | createAccount | bond | name | proposal | input | batch | identify | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole"
				}
			}
		},
		{
			"Address": "1815E3667F406CA3872234D2573014CDE6CD2ABC",
			"PublicKey": null,
			"Amount": 378,
			"Name": "Lala",
			"Permissions": {
				"Base": {
					"Perms": "root | send | call | createContract | createAccount | bond | name | proposal | input | batch | identify | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole",
					"SetBit": "root | send | call | createContract | createAccount | bond | name | proposal | input | batch | identify | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole"
				}
			}
		},
		{
			"Address": "A8174022832E4BA1BB52B380128514F72FB2FEBD",
			"PublicKey": null,
			"Amount": 191,
			"Name": "Po",
			"Permissions": {
				"Base": {
					"Perms": "root | send | call | createContract | createAccount | bond | name | proposal | input | batch | identify | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole",
					"SetBit": "root | send | call | createContract | createAccount | bond | name | proposal | input | batch | identify | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole"
				}
			}
		},
		{
			"Address": "8979C634DFD2F6D20975EBE02C34B5E9C280AB18",
			"PublicKey": null,
			"Amount": 1304,
			"Name": "Tinkie-winkie",
			"Permissions": {
				"Base": {
					"Perms": "root | send | call | createContract | createAccount | bond | name | proposal | input | batch | identify | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole",
					"SetBit": "root | send | call | createContract | createAccount | bond | name | proposal | input | batch | identify | hasBase | setBase | unsetBase | setGlobal | hasRole | addRole | removeRole"
				}
			}
		}
	],
	"Validators": [
		{
			"Address": "29BB63AD75E2DA3368FC823DC68C01CF1FA87190",
			"PublicKey": {
				"CurveType": "ed25519",
				"PublicKey": "A18679ADC4391630178AC0DB35A115BEAEDA38B3DEABF92AC5FDE31A748DC259"
			},
			"Amount": 277,
			"Name": "Bar",
			"UnbondTo": [
				{
					"Address": "29BB63AD75E2DA3368FC823DC68C01CF1FA87190",
					"PublicKey": null,
					"Amount": 277
				}
			]
		},
		{
			"Address": "C92303227C9F0EC569B27B02DB328CFA0A7DF7E0",
			"PublicKey": {
				"CurveType": "ed25519",
				"PublicKey": "E3C56A2C047C9C82036778620E6F9089E5FB38A5D36CE47D9545CBA930C79522"
			},
			"Amount": 285,
			"Name": "Baz",
			"UnbondTo": [
				{
					"Address": "C92303227C9F0EC569B27B02DB328CFA0A7DF7E0",
					"PublicKey": null,
					"Amount": 285
				}
			]
		},
		{
			"Address": "900EBED8C6B27F7B606B6CAD34DB03C2C5C0E541",
			"PublicKey": {
				"CurveType": "ed25519",
				"PublicKey": "78133DE07C66C616263FD2D6A54BD3FC0D8CBF3A87BB2E0C410A0C8DEB6189CE"
			},
			"Amount": 292,
			"Name": "Foo",
			"UnbondTo": [
				{
					"Address": "900EBED8C6B27F7B606B6CAD34DB03C2C5C0E541",
					"PublicKey": null,
					"Amount": 292
				}
			]
		}
	]
}`
