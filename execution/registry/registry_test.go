package registry

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeProtobuf(t *testing.T) {
	entry := &RegisteredNode{
		Moniker: "test",
		ID:      crypto.Address{1, 2, 3, 4, 5},
		PublicKey: crypto.PublicKey{
			CurveType: crypto.CurveTypeEd25519,
			PublicKey: binary.HexBytes{1, 2, 3, 4, 5},
		},
		NetAddress: "localhost",
	}
	encoded, err := proto.Marshal(entry)
	require.NoError(t, err)
	entryOut := new(RegisteredNode)
	err = proto.Unmarshal(encoded, entryOut)
	require.NoError(t, err)
	assert.Equal(t, entry, entryOut)
}
