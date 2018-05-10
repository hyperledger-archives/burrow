package account

import (
	"bytes"
	"testing"

	"github.com/hyperledger/burrow/util/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// From go-ethereum/crypto/signature_test.go
var (
	testmsg    = hexutil.MustDecode("0xce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008")
	testsig    = hexutil.MustDecode("0x90f27b8b488db00b00606796d2987f6a5f59ae62ea05effe84fef5b8b0e549984a691139ad57a3f0b906637673aa2f63d1f55cb1a69199d4009eea23ceaddc9301")
	testpubkey = hexutil.MustDecode("0x04e32df42865e97135acfb65f3bae71bdc86f4d49150ad6a440b6f15878109880a0a2b2667f7e725ceea70c673093bf67663e0312623c8e091b13cf2c0f11ef652")
)

func TestGeneratePrivateKey(t *testing.T) {
	privateKey, err := GeneratePrivateKey(bytes.NewBuffer([]byte{
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
	}))
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

// From go-ethereum/crypto/signature_test.go
func TestEcrecover(t *testing.T) {
	pubkey, err := EcRecover(testmsg, testsig)
	if err != nil {
		t.Fatalf("recover error: %s", err)
	}
	if !bytes.Equal(pubkey, testpubkey) {
		t.Errorf("pubkey mismatch: want: %x have: %x", testpubkey, pubkey)
	}
}
