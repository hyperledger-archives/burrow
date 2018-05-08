package deployment

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockKeyClient_DumpKeys(t *testing.T) {
	keyClient := mock.NewKeyClient()
	_, err := keyClient.Generate("foo", crypto.CurveTypeEd25519)
	require.NoError(t, err)
	_, err = keyClient.Generate("foobar", crypto.CurveTypeEd25519)
	require.NoError(t, err)
	pkg := Package{Keys: keyClient.Keys()}
	dump, err := pkg.Dump(DefaultDumpKeysFormat)
	require.NoError(t, err)

	// Check JSON equal
	var keys struct{ Keys []*mock.Key }
	err = json.Unmarshal([]byte(dump), &keys)
	require.NoError(t, err)
	bs, err := json.MarshalIndent(keys, "", "  ")
	require.NoError(t, err)
	assert.Equal(t, string(bs), dump)
}

func TestMockKeyClient_DumpKeysKubernetes(t *testing.T) {
	keyClient := mock.NewKeyClient()
	_, err := keyClient.Generate("foo", crypto.CurveTypeEd25519)
	require.NoError(t, err)
	_, err = keyClient.Generate("foobar", crypto.CurveTypeEd25519)
	require.NoError(t, err)
	pkg := Package{Keys: keyClient.Keys()}
	dump, err := pkg.Dump(KubernetesKeyDumpFormat)
	require.NoError(t, err)
	fmt.Println(dump)
}

func TestMockKeyClient_DumpKeysHelm(t *testing.T) {
	keyClient := mock.NewKeyClient()
	_, err := keyClient.Generate("foo", crypto.CurveTypeEd25519)
	require.NoError(t, err)
	_, err = keyClient.Generate("foobar", crypto.CurveTypeEd25519)
	require.NoError(t, err)
	pkg := Package{Keys: keyClient.Keys()}
	dump, err := pkg.Dump(HelmDumpKeysFormat)
	require.NoError(t, err)
	fmt.Println(dump)
}
