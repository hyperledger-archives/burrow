// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package structure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValuesAndContext(t *testing.T) {
	keyvals := []interface{}{"hello", 1, "dog", 2, "fish", 3, "fork", 5}
	vals, ctx := ValuesAndContext(keyvals, "hello", "fish")
	assert.Equal(t, map[string]interface{}{"hello": 1, "fish": 3}, vals)
	assert.Equal(t, []interface{}{"dog", 2, "fork", 5}, ctx)
}

func TestKeyValuesMap(t *testing.T) {
	keyvals := []interface{}{
		[][]interface{}{{2}}, 3,
		"hello", 1,
		"fish", 3,
		"dog", 2,
		"fork", 5,
	}
	vals := KeyValuesMap(keyvals)
	assert.Equal(t, map[string]interface{}{
		"[[2]]": 3,
		"hello": 1,
		"fish":  3,
		"dog":   2,
		"fork":  5,
	}, vals)
}

func TestVectorise(t *testing.T) {
	kvs := []interface{}{
		"scope", "lawnmower",
		"hub", "budub",
		"occupation", "fish brewer",
		"scope", "hose pipe",
		"flub", "dub",
		"scope", "rake",
		"flub", "brub",
	}

	kvsVector := Vectorise(kvs, "occupation", "scope")
	// Vectorise scope
	assert.Equal(t, []interface{}{
		"scope", Vector{"lawnmower", "hose pipe", "rake"},
		"hub", "budub",
		"occupation", "fish brewer",
		"flub", Vector{"dub", "brub"},
	},
		kvsVector)
}

func TestVector_String(t *testing.T) {
	vec := Vector{"one", "two", "grue"}
	assert.Equal(t, "[one two grue]", vec.String())
}

func TestRemoveKeys(t *testing.T) {
	// Remove multiple of same key
	assert.Equal(t, []interface{}{"Fish", 9},
		RemoveKeys([]interface{}{"Foo", "Bar", "Fish", 9, "Foo", "Baz", "odd-key"},
			"Foo"))

	// Remove multiple different keys
	assert.Equal(t, []interface{}{"Fish", 9},
		RemoveKeys([]interface{}{"Foo", "Bar", "Fish", 9, "Foo", "Baz", "Bar", 89},
			"Foo", "Bar"))

	// Remove nothing but supply keys
	assert.Equal(t, []interface{}{"Foo", "Bar", "Fish", 9},
		RemoveKeys([]interface{}{"Foo", "Bar", "Fish", 9},
			"A", "B", "C"))

	// Remove nothing since no keys supplied
	assert.Equal(t, []interface{}{"Foo", "Bar", "Fish", 9},
		RemoveKeys([]interface{}{"Foo", "Bar", "Fish", 9}))
}

func TestDelete(t *testing.T) {
	assert.Equal(t, []interface{}{1, 2, 4, 5}, Delete([]interface{}{1, 2, 3, 4, 5}, 2, 1))
}

func TestCopyPrepend(t *testing.T) {
	assert.Equal(t, []interface{}{"three", 4, 1, "two"},
		CopyPrepend([]interface{}{1, "two"}, "three", 4))
	assert.Equal(t, []interface{}{}, CopyPrepend(nil))
	assert.Equal(t, []interface{}{1}, CopyPrepend(nil, 1))
	assert.Equal(t, []interface{}{1}, CopyPrepend([]interface{}{1}))
}
