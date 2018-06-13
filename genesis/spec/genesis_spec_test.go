package spec

import (
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys/mock"
	permission "github.com/hyperledger/burrow/permission/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesisSpec_GenesisDoc(t *testing.T) {
	keyClient := mock.NewKeyClient()

	// Try a spec with a single account/validator
	amtBonded := uint64(100)
	genesisSpec := GenesisSpec{
		Accounts: []TemplateAccount{{
			AmountBonded: &amtBonded,
		}},
	}

	genesisDoc, err := genesisSpec.GenesisDoc(keyClient)
	require.NoError(t, err)
	require.Len(t, genesisDoc.Accounts, 1)
	// Should create validator
	require.Len(t, genesisDoc.Validators, 1)
	assert.NotZero(t, genesisDoc.Accounts[0].Address)
	assert.NotZero(t, genesisDoc.Accounts[0].PublicKey)
	assert.Equal(t, genesisDoc.Accounts[0].Address, genesisDoc.Validators[0].Address)
	assert.Equal(t, genesisDoc.Accounts[0].PublicKey, genesisDoc.Validators[0].PublicKey)
	assert.Equal(t, amtBonded, genesisDoc.Validators[0].Amount)
	assert.NotEmpty(t, genesisDoc.ChainName, "Chain name should not be empty")

	address, err := keyClient.Generate("test-lookup-of-key", crypto.CurveTypeEd25519)
	require.NoError(t, err)
	pubKey, err := keyClient.PublicKey(address)
	require.NoError(t, err)

	// Try a spec with two accounts and no validators
	amt := uint64(99299299)
	genesisSpec = GenesisSpec{
		Accounts: []TemplateAccount{
			{
				Address: &address,
			},
			{
				Amount:      &amt,
				Permissions: []string{permission.CreateAccountString, permission.CallString},
			}},
	}

	genesisDoc, err = genesisSpec.GenesisDoc(keyClient)
	require.NoError(t, err)

	require.Len(t, genesisDoc.Accounts, 2)
	// Nothing bonded so no validators
	require.Len(t, genesisDoc.Validators, 0)
	assert.Equal(t, pubKey, genesisDoc.Accounts[0].PublicKey)
	assert.Equal(t, amt, genesisDoc.Accounts[1].Amount)
	permFlag := permission.CreateAccount | permission.Call
	assert.Equal(t, permFlag, genesisDoc.Accounts[1].Permissions.Base.Perms)
	assert.Equal(t, permFlag, genesisDoc.Accounts[1].Permissions.Base.SetBit)

	// Try an empty spec
	genesisSpec = GenesisSpec{}

	genesisDoc, err = genesisSpec.GenesisDoc(keyClient)
	require.NoError(t, err)

	// Similar assersions to first case - should generate our default single identity chain
	require.Len(t, genesisDoc.Accounts, 1)
	// Should create validator
	require.Len(t, genesisDoc.Validators, 1)
	assert.NotZero(t, genesisDoc.Accounts[0].Address)
	assert.NotZero(t, genesisDoc.Accounts[0].PublicKey)
	assert.Equal(t, genesisDoc.Accounts[0].Address, genesisDoc.Validators[0].Address)
	assert.Equal(t, genesisDoc.Accounts[0].PublicKey, genesisDoc.Validators[0].PublicKey)
}

func TestTemplateAccount_AccountPermissions(t *testing.T) {
}
