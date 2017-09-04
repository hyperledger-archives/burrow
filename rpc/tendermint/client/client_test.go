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

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParamsMap(t *testing.T) {
	type aStruct struct {
		Baz int
	}
	dict, err := paramsMap("Foo", aStruct{5},
		"Bar", "Nibbles")
	assert.NoError(t, err, "Should not be a paramsMaperror")
	assert.Equal(t, map[string]interface{}{
		"Foo": aStruct{5},
		"Bar": "Nibbles",
	}, dict)

	// Empty map
	dict, err = paramsMap()
	assert.Equal(t, map[string]interface{}{}, dict)
	assert.NoError(t, err, "Empty mapsAndValues call should be fine")

	// Invalid maps
	assert.NoError(t, err, "Empty mapsAndValues call should be fine")
	_, err = paramsMap("Foo", 4, "Bar")
	assert.Error(t, err, "Should be an error to get an odd number of arguments")

	_, err = paramsMap("Foo", 4, 4, "Bar")
	assert.Error(t, err, "Should be an error to provide non-string keys")
}
