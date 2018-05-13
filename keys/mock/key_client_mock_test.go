package mock

import (
	"testing"

	"encoding/json"

	"fmt"

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

func TestMockKeyClient_DumpKeysKubernetes(t *testing.T) {
	keyClient := NewMockKeyClient()
	_, err := keyClient.Generate("foo", keys.KeyTypeEd25519Ripemd160)
	require.NoError(t, err)
	_, err = keyClient.Generate("foobar", keys.KeyTypeEd25519Ripemd160)
	require.NoError(t, err)
	dump, err := keyClient.DumpKeys(KubernetesKeyDumpFormat)
	require.NoError(t, err)
	fmt.Println(dump)
}

func TestMockKey_MonaxKeyJSON(t *testing.T) {
	key, err := newMockKey("monax-key-test")
	require.NoError(t, err)
	monaxKey := key.MonaxKeyJSON()
	t.Logf("key is: %v", monaxKey)
	keyJSON := &plainKeyJSON{}
	err = json.Unmarshal([]byte(monaxKey), keyJSON)
	require.NoError(t, err)
	// byte length of UUID string = 16 * 2 + 4 = 36
	assert.Len(t, keyJSON.Id, 36)
	assert.Equal(t, key.Address.String(), keyJSON.Address)
	assert.Equal(t, key.PrivateKey, keyJSON.PrivateKey)
	assert.Equal(t, string(keys.KeyTypeEd25519Ripemd160), keyJSON.Type)
}
