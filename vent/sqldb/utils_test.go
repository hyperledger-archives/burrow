package sqldb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDigits(t *testing.T) {
	s := fmt.Sprintf("%v", maxUint64)
	assert.Len(t, s, digits(maxUint64))
	assert.Equal(t, 1, digits(1))
	assert.Equal(t, 1, digits(1))
	assert.Equal(t, 1, digits(2))
	assert.Equal(t, 2, digits(10))
}
