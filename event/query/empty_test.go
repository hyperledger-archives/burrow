package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyQueryMatchesAnything(t *testing.T) {
	q := Empty{}
	assert.True(t, q.Matches(TagMap{}))
	assert.True(t, q.Matches(TagMap{"Asher": "Roth"}))
	assert.True(t, q.Matches(TagMap{"Route": "66"}))
	assert.True(t, q.Matches(TagMap{"Route": "66", "Billy": "Blue"}))
}
