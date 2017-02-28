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

package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyAppend(t *testing.T) {
	assert.Equal(t, Slice(1, "two", "three", 4),
		CopyAppend(Slice(1, "two"), "three", 4))
	assert.Equal(t, EmptySlice(), CopyAppend(nil))
	assert.Equal(t, Slice(1), CopyAppend(nil, 1))
	assert.Equal(t, Slice(1), CopyAppend(Slice(1)))
}

func TestCopyPrepend(t *testing.T) {
	assert.Equal(t, Slice("three", 4, 1, "two"),
		CopyPrepend(Slice(1, "two"), "three", 4))
	assert.Equal(t, EmptySlice(), CopyPrepend(nil))
	assert.Equal(t, Slice(1), CopyPrepend(nil, 1))
	assert.Equal(t, Slice(1), CopyPrepend(Slice(1)))
}

func TestConcat(t *testing.T) {
	assert.Equal(t, Slice(1, 2, 3, 4, 5), Concat(Slice(1, 2, 3, 4, 5)))
	assert.Equal(t, Slice(1, 2, 3, 4, 5), Concat(Slice(1, 2, 3), Slice(4, 5)))
	assert.Equal(t, Slice(1, 2, 3, 4, 5), Concat(Slice(1), Slice(2, 3), Slice(4, 5)))
	assert.Equal(t, EmptySlice(), Concat(nil))
	assert.Equal(t, Slice(1), Concat(nil, Slice(), Slice(1)))
	assert.Equal(t, Slice(1), Concat(Slice(1), Slice(), nil))
}

func TestDelete(t *testing.T) {
	assert.Equal(t, Slice(1, 2, 4, 5), Delete(Slice(1, 2, 3, 4, 5), 2, 1))
}

func TestDeepFlatten(t *testing.T) {
	assert.Equal(t, Flatten(Slice(Slice(1, 2), 3, 4)), Slice(1, 2, 3, 4))
	nestedSlice := Slice(Slice(1, Slice(Slice(2))), Slice(3, 4))
	assert.Equal(t, DeepFlatten(nestedSlice, -1), Slice(1, 2, 3, 4))
	assert.Equal(t, DeepFlatten(nestedSlice, 2), Slice(1, Slice(2), 3, 4))
}
