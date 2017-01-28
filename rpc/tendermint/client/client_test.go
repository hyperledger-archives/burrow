package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapsAndValues(t *testing.T) {
	type aStruct struct {
		Baz int
	}
	dict, vals, err := mapAndValues("Foo", aStruct{5},
		"Bar", "Nibbles")
	assert.Equal(t, map[string]interface{}{
		"Foo": aStruct{5},
		"Bar": "Nibbles",
	}, dict)
	assert.Equal(t, []interface{}{aStruct{5}, "Nibbles"}, vals)

	// Empty map
	dict, vals, err = mapAndValues()
	assert.Equal(t, map[string]interface{}{}, dict)
	assert.Equal(t, []interface{}{}, vals)
	assert.NoError(t, err, "Empty mapsAndValues call should be fine")

	// Invalid maps
	assert.NoError(t, err, "Empty mapsAndValues call should be fine")
	_, _, err = mapAndValues("Foo", 4, "Bar")
	assert.Error(t, err, "Should be an error to get an odd number of arguments")

	_, _, err = mapAndValues("Foo", 4, 4, "Bar")
	assert.Error(t, err, "Should be an error to provide non-string keys")
}
