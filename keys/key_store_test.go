package keys

import (
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyStore(t *testing.T) {
	ks, cleanup := EnterTestKeyStore()
	defer cleanup()

	kED, err := ks.Gen("", crypto.CurveTypeEd25519)
	require.NoError(t, err)
	assert.Equal(t, kED.CurveType, crypto.CurveTypeEd25519)

	kSECP, err := ks.Gen("", crypto.CurveTypeSecp256k1)
	require.NoError(t, err)
	assert.Equal(t, kSECP.CurveType, crypto.CurveTypeSecp256k1)

	err = ks.AddName("foo", kED.Address)
	require.NoError(t, err)

	address, err := ks.GetName("foo")
	require.NoError(t, err)
	assert.Equal(t, kED.Address, address)

	err = ks.AddName("bar", kSECP.Address)
	require.NoError(t, err)

	address, err = ks.GetName("bar")
	require.NoError(t, err)
	assert.Equal(t, kSECP.Address, address)

	adds, err := ks.GetAllAddresses()
	require.NoError(t, err)
	assert.ElementsMatch(t, adds, []crypto.Address{kED.Address, kSECP.Address})

	names, err := ks.GetAllNames()
	require.NoError(t, err)
	assert.Equal(t, len(names), 2)
	assert.Equal(t, names["foo"], kED.Address)
	assert.Equal(t, names["bar"], kSECP.Address)

	err = ks.RmName("foo")
	require.NoError(t, err)
	err = ks.RmName("bar")
	require.NoError(t, err)

	names, err = ks.GetAllNames()
	require.NoError(t, err)
	assert.ElementsMatch(t, names, []string{})
	assert.Equal(t, len(names), 0)
}
