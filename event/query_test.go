package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder()
	qry, err := qb.Query()
	require.NoError(t, err)
	assert.Equal(t, emptyString, qry.String())

	qb = qb.AndGreaterThanOrEqual("foo.size", 45)
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.Equal(t, "foo.size >= 45", qry.String())

	qb = qb.AndEquals("bar.name", "marmot")
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.Equal(t, "foo.size >= 45 AND bar.name = 'marmot'", qry.String())

	assert.True(t, qry.Matches(map[string]interface{}{"foo.size": 80, "bar.name": "marmot"}))
	assert.False(t, qry.Matches(map[string]interface{}{"foo.size": 8, "bar.name": "marmot"}))
	assert.False(t, qry.Matches(map[string]interface{}{"foo.size": 80, "bar.name": "marot"}))

	qb = qb.AndContains("bar.desc", "burrow")
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.Equal(t, "foo.size >= 45 AND bar.name = 'marmot' AND bar.desc CONTAINS 'burrow'", qry.String())

	assert.True(t, qry.Matches(map[string]interface{}{"foo.size": 80, "bar.name": "marmot", "bar.desc": "lives in a burrow"}))
	assert.False(t, qry.Matches(map[string]interface{}{"foo.size": 80, "bar.name": "marmot", "bar.desc": "lives in a shoe"}))

	qb = NewQueryBuilder().AndEquals("foo", "bar")
	qb = qb.And(NewQueryBuilder().AndGreaterThanOrEqual("frogs", 4))
	qry, err = qb.Query()
	require.NoError(t, err)
	assert.Equal(t, "foo = 'bar' AND frogs >= 4", qry.String())
}
