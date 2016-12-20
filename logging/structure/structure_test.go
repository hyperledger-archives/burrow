package structure

import (
	"testing"

	. "github.com/eris-ltd/eris-db/util/slice"
	"github.com/stretchr/testify/assert"
)

func TestValuesAndContext(t *testing.T) {
	keyvals := Slice("hello", 1, "dog", 2, "fish", 3, "fork", 5)
	vals, ctx := ValuesAndContext(keyvals, "hello", "fish")
	assert.Equal(t, map[interface{}]interface{}{"hello": 1, "fish": 3}, vals)
	assert.Equal(t, Slice("dog", 2, "fork", 5), ctx)
}

func TestVectorise(t *testing.T) {
	kvs := Slice(
		"scope", "lawnmower",
		"hub", "budub",
		"occupation", "fish brewer",
		"scope", "hose pipe",
		"flub", "dub",
		"scope", "rake",
		"flub", "brub",
	)

	kvsVector := Vectorise(kvs, "occupation", "scope")
	assert.Equal(t, Slice(
		"scope", Slice("lawnmower", "hose pipe", "rake"),
		"hub", "budub",
		"occupation", "fish brewer",
		"flub", Slice("dub", "brub"),
	),
		kvsVector)
}
