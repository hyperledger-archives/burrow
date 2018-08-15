package infoclient

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
