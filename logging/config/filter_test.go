package config

import (
	"testing"

	"github.com/eris-ltd/eris-db/logging/config/types"
	. "github.com/eris-ltd/eris-db/util/slice"
	"github.com/stretchr/testify/assert"
)

func TestBuildKeyValuesPredicateMatchAll(t *testing.T) {
	conf := []*types.KeyValuePredicateConfig{
		{
			KeyRegex: "Foo",
			ValueRegex: "bar",
		},
	}
	kvp, err := BuildKeyValuesPredicate(conf, true)
	assert.NoError(t, err)
	assert.True(t, kvp(Slice("Foo", "bar", "Bosh", "Bish")))
}

func TestBuildKeyValuesPredicateMatchAny(t *testing.T) {
	conf := []*types.KeyValuePredicateConfig{
		{
			KeyRegex:   "Bosh",
			ValueRegex: "Bish",
		},
	}
	kvp, err := BuildKeyValuesPredicate(conf, false)
	assert.NoError(t, err)
	assert.True(t, kvp(Slice("Foo", "bar", "Bosh", "Bish")))
}

func TestBuildFilterPredicate(t *testing.T) {
	fc := &types.FilterConfig{
		Include: []*types.KeyValuePredicateConfig{
			{
				KeyRegex: "^Foo$",
			},
		},
		Exclude: []*types.KeyValuePredicateConfig{
			{
				KeyRegex:   "Bosh",
				ValueRegex: "Bish",
			},
		},
	}
	fp, err := BuildFilterPredicate(fc)
	assert.NoError(t, err)
	assert.False(t, fp(Slice("Foo", "bar", "Shoes", 42)))
	assert.True(t, fp(Slice("Foo", "bar", "Shoes", 42, "Bosh", "Bish")))
	assert.True(t, fp(Slice("Food", 0.2, "Shoes", 42)))
}
