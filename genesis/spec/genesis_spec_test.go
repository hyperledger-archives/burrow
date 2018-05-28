package spec

import (
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/permission"
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
	require.Len(t, genesisDoc.Accounts(), 1)
	// Should create validator
	require.Len(t, genesisDoc.Validators(), 1)
	assert.NotZero(t, genesisDoc.Accounts()[0].Address())
	assert.NotZero(t, genesisDoc.Accounts()[0].PublicKey())
	assert.Equal(t, genesisDoc.Accounts()[0].Address(), genesisDoc.Validators()[0].Address())
	assert.Equal(t, genesisDoc.Accounts()[0].PublicKey(), genesisDoc.Validators()[0].PublicKey())
	assert.Equal(t, amtBonded, genesisDoc.Validators()[0].Stake())
	assert.NotEmpty(t, genesisDoc.ChainName, "Chain name should not be empty")

	address, err := keyClient.Generate("test-lookup-of-key", keys.KeyTypeEd25519Ripemd160)
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

	require.Len(t, genesisDoc.Accounts(), 2)
	// Nothing bonded so no validators
	require.Len(t, genesisDoc.Validators(), 0)
	acc0 := genesisDoc.Accounts()[0]
	acc1 := genesisDoc.Accounts()[1]
	var acc acm.Account

	// genesis sorts account by their address!
	if acc0.PublicKey() == pubKey {
		acc = acc1
		assert.Equal(t, pubKey, acc0.PublicKey())
	} else {
		acc = acc0
		assert.Equal(t, pubKey, acc1.PublicKey())
	}

	assert.Equal(t, amt, acc.Balance())
	permFlag := permission.CreateAccount | permission.Call
	assert.Equal(t, permFlag, acc.Permissions().Base.Perms)
	assert.Equal(t, permFlag, acc.Permissions().Base.SetBit)

	// Try an empty spec
	genesisSpec = GenesisSpec{}

	genesisDoc, err = genesisSpec.GenesisDoc(keyClient)
	require.NoError(t, err)

	// Similar assersions to first case - should generate our default single identity chain
	require.Len(t, genesisDoc.Accounts(), 1)
	// Should create validator
	require.Len(t, genesisDoc.Validators(), 1)
	assert.NotZero(t, genesisDoc.Accounts()[0].Address())
	assert.NotZero(t, genesisDoc.Accounts()[0].PublicKey())
	assert.Equal(t, genesisDoc.Accounts()[0].Address(), genesisDoc.Validators()[0].Address())
	assert.Equal(t, genesisDoc.Accounts()[0].PublicKey(), genesisDoc.Validators()[0].PublicKey())
}

func TestTemplateAccount_AccountPermissions(t *testing.T) {
}
