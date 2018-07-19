package crypto

import (
	"encoding/json"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicKeySerialisation(t *testing.T) {
	priv := PrivateKeyFromSecret("foo", CurveTypeEd25519)
	pub := priv.GetPublicKey()
	expectedAddress := Address{
		0x83, 0x20, 0x78, 0x17, 0xdc, 0x38, 0x14, 0xb9, 0x6f, 0x57,
		0xef, 0xf9, 0x25, 0xf4, 0x67, 0xe0, 0x7c, 0xaa, 0x91, 0x38,
	}
	assert.Equal(t, expectedAddress, pub.Address())
	bs, err := proto.Marshal(&pub)
	require.NoError(t, err)
	var pubOut PublicKey
	err = proto.Unmarshal(bs, &pubOut)
	assert.Equal(t, pub, pubOut)

	bs, err = json.Marshal(pub)
	require.NoError(t, err)
	assert.Equal(t, `{"CurveType":"ed25519","PublicKey":"34D26579DBB456693E540672CF922F52DDE0D6532E35BF06BE013A7C532F20E0"}`,
		string(bs))
}
