package mock

import (
	"testing"

	"encoding/json"

	"github.com/hyperledger/burrow/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockKeyClient_DumpKeys(t *testing.T) {
	keyClient := NewMockKeyClient()
	_, err := keyClient.Generate("foo", keys.KeyTypeEd25519Ripemd160)
	require.NoError(t, err)
	_, err = keyClient.Generate("foobar", keys.KeyTypeEd25519Ripemd160)
	require.NoError(t, err)
	dump, err := keyClient.DumpKeys(DefaultDumpKeysFormat)
	require.NoError(t, err)

	// Check JSON equal
	var keys struct{ Keys []*MockKey }
	err = json.Unmarshal([]byte(dump), &keys)
	require.NoError(t, err)
	bs, err := json.MarshalIndent(keys, "", "  ")
	require.NoError(t, err)
	assert.Equal(t, string(bs), dump)
}
