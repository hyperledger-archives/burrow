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
		0x4, 0x5f, 0x56, 0x0, 0x65, 0x41, 0x82, 0xcf, 0xea, 0xcc,
		0xfe, 0x6c, 0xb1, 0x9f, 0x6, 0x42, 0xe8, 0xa5, 0x98, 0x98,
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
