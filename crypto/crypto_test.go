package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePrivateKey(t *testing.T) {
	privateKey, err := GeneratePrivateKey(bytes.NewBuffer([]byte{
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
	}), CurveTypeEd25519)
	require.NoError(t, err)
	assert.NoError(t, EnsureEd25519PrivateKeyCorrect(privateKey.RawBytes()))
	badKey := privateKey.RawBytes()
	// Change part of the public part to not match private part
	badKey[35] = 2
	assert.Error(t, EnsureEd25519PrivateKeyCorrect(badKey))
	goodKey := privateKey.RawBytes()
	// Change part of the private part invalidating public part
	goodKey[31] = 2
	assert.Error(t, EnsureEd25519PrivateKeyCorrect(badKey))
}
