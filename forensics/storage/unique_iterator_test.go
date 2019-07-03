package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueIterator(t *testing.T) {
	it := iteratorOver(kvPairs("a", "dogs", "a", "pogs",
		"b", "slime", "b", "grime", "b", "nogs",
		"c", "strudel"))

	assert.Equal(t, kvPairs("a", "dogs", "b", "slime", "c", "strudel"),
		collectIterator(Uniq(it)))
}
