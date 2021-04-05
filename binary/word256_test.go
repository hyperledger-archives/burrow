package binary

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmthrgd/go-hex"
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

func TestTwosComplement(t *testing.T) {
	v := Int64ToWord256(-10)
	require.Equal(t, "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF6",
		hex.EncodeUpperToString(v[:]))

	upper, ok := new(big.Int).SetString("32423973453453434237423", 10)
	require.True(t, ok)
	inc, ok := new(big.Int).SetString("3242397345345343421", 10)
	require.True(t, ok)
	for i := new(big.Int).Neg(upper); i.Cmp(upper) == -1; i.Add(i, inc) {
		v := BigIntFromWord256(BigIntToWord256(i))
		require.True(t, i.Cmp(v) == 0, "expected %d == %d", i, v)
	}
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
	require.NoError(t, err)
	assert.Equal(t, w, *bs2)
}

func TestInt64ToWord256(t *testing.T) {
	i := int64(-34)
	assert.Equal(t, i, Int64FromWord256(Int64ToWord256(i)))
}
