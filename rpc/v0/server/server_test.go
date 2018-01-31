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

package server

import (
	//"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit tests for server components goes here. Full-on client-server tests
// can be found in the test folder.

func TestIdGet(t *testing.T) {
	idPool := NewIdPool(100)
	idparr := make([]uint, 100)
	arr := make([]uint, 100)
	for i := 0; i < 100; i++ {
		idparr[i] = uint(i + 1)
		arr[i], _ = idPool.GetId()
	}
	assert.Equal(t, idparr, arr, "Array of gotten id's is not [1, 2, ..., 101] as expected")
}

func TestIdPut(t *testing.T) {
	idPool := NewIdPool(10)
	for i := 0; i < 10; i++ {
		idPool.GetId()
	}
	idPool.ReleaseId(5)
	id, _ := idPool.GetId()
	assert.Equal(t, id, uint(5), "Id gotten is not 5.")
}

func TestIdFull(t *testing.T) {
	idPool := NewIdPool(10)
	for i := 0; i < 10; i++ {
		idPool.GetId()
	}
	_, err := idPool.GetId()
	assert.Error(t, err)
}
