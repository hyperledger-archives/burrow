package mock

import (
	"testing"

	"encoding/json"

	"github.com/hyperledger/burrow/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockKey_MonaxKeyJSON(t *testing.T) {
	key, err := newKey("monax-key-test")
	require.NoError(t, err)
	monaxKey := key.MonaxKeysJSON()
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
