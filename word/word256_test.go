package word

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
