package binary

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"encoding/json"
)

func TestWord256_UnpadLeft(t *testing.T) {
	bs := []byte{0x45, 0x12}
	w := LeftPadWord256(bs)
	wExpected := Word256{}
	wExpected[30] = bs[0]
	wExpected[31] = bs[1]
	assert.Equal(t, wExpected, w)
	assert.Equal(t, bs, w.UnpadLeft())
}

func TestWord256_UnpadRight(t *testing.T) {
	bs := []byte{0x45, 0x12}
	w := RightPadWord256(bs)
	wExpected := Word256{}
	wExpected[0] = bs[0]
	wExpected[1] = bs[1]
	assert.Equal(t, wExpected, w)
	assert.Equal(t, bs, w.UnpadRight())
}

func TestLeftPadWord256(t *testing.T) {
	assert.Equal(t, Zero256, LeftPadWord256(nil))
	assert.Equal(t,
		Word256{
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 1, 2, 3,
		},
		LeftPadWord256([]byte{1, 2, 3}))
}

func TestOne256(t *testing.T) {
	assert.Equal(t, Int64ToWord256(1), One256)
}

func TestWord256_MarshalText(t *testing.T) {
	w := Word256{1, 2, 3, 4, 5}
	out, err := json.Marshal(w)
	require.NoError(t, err)
	assert.Equal(t, "\"0102030405000000000000000000000000000000000000000000000000000000\"", string(out))
	bs2 := new(Word256)
	err = json.Unmarshal(out, bs2)
	assert.Equal(t, w, *bs2)
}
