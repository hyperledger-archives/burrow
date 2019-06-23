package binary

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHexBytes_MarshalText(t *testing.T) {
	bs := HexBytes{1, 2, 3, 4, 5}
	out, err := json.Marshal(bs)
	require.NoError(t, err)
	assert.Equal(t, "\"0102030405\"", string(out))
	bs2 := new(HexBytes)
	err = json.Unmarshal(out, bs2)
	require.NoError(t, err)
	assert.Equal(t, bs, *bs2)
}
