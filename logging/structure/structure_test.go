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

	. "github.com/monax/eris-db/util/slice"
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
