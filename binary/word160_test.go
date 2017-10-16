package binary

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWord160_Word256(t *testing.T) {
	word256 := Word256{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		1, 2, 3, 4, 5,
		6, 7, 8, 9, 10,
		11, 12, 13, 14, 15,
		16, 17, 18, 19, 20,
	}
	word160 := Word160{
		1, 2, 3, 4, 5,
		6, 7, 8, 9, 10,
		11, 12, 13, 14, 15,
		16, 17, 18, 19, 20,
	}
	assert.Equal(t, word256, word160.Word256())
	assert.Equal(t, word160, word256.Word160())
}
