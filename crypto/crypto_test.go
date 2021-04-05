package crypto

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
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

func TestSigning(t *testing.T) {
	for _, ct := range []CurveType{CurveTypeSecp256k1, CurveTypeEd25519} {
		t.Run(fmt.Sprintf("%v signing", ct), func(t *testing.T) {
			pk, err := GeneratePrivateKey(rand.Reader, ct)
			require.NoError(t, err)
			msg := []byte(("Flipity flobity floo"))
			sig, err := pk.Sign(msg)
			require.NoError(t, err)
			err = pk.GetPublicKey().Verify(msg, sig)
			require.NoError(t, err)
		})

	}

	t.Run("EthSignature", func(t *testing.T) {
		pk := PrivateKeyFromSecret("seee", CurveTypeSecp256k1)
		msg := []byte(("Flipity flobity floo"))
		sig, err := pk.Sign(msg)
		require.NoError(t, err)
		ethSig, err := sig.GetEthSignature(big.NewInt(12342))
		require.NoError(t, err)
		parity := ethSig.RecoveryIndex()
		require.True(t, parity == 0 || parity == 1)

		// Now verify signature comes out intact after serialisation
		compactSig, err := ethSig.ToCompactSignature()
		require.NoError(t, err)
		sigOut, err := SignatureFromBytes(compactSig, CurveTypeSecp256k1)
		require.NoError(t, err)
		err = pk.GetPublicKey().Verify(msg, sigOut)
		require.NoError(t, err)
	})
}
