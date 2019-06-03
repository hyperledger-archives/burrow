// Copyright 2019 Monax Industries Limited
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

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareKeys(t *testing.T) {
	assert.Equal(t, 1, CompareKeys(nil, []byte{2}))
	assert.Equal(t, -1, CompareKeys([]byte{2}, nil))
	assert.Equal(t, -1, CompareKeys([]byte{}, nil))
	assert.Equal(t, 1, CompareKeys(nil, []byte{}))
	assert.Equal(t, 0, CompareKeys(nil, nil))
	assert.Equal(t, -1, CompareKeys([]byte{1, 2, 3}, []byte{2}))
}
