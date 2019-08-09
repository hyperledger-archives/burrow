package abi

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnpackEvent(t *testing.T) {
	eventAbi := `{"anonymous":false,"name":"TestEvent","type":"event","inputs":[{"indexed":true,"name":"direction","type":"bytes32"},{"indexed":false,"name":"trueism","type":"bool"},{"indexed":true,"name":"newDepth","type":"int128"},{"indexed":true,"name":"hash","type":"string"}]}`

	type args struct {
		Direction string
		Trueism   bool
		NewDepth  int64
		Hash      string
	}
	in := &args{
		Direction: "foo",
		Trueism:   true,
		NewDepth:  232,
		Hash:      "DEADBEEFCAFEBADE01234567DEADBEEF",
	}
	eventSpec := new(EventSpec)

	err := json.Unmarshal([]byte(eventAbi), eventSpec)
	require.NoError(t, err)

	topics, data, err := PackEvent(eventSpec, in)
	require.NoError(t, err)

	out := new(args)
	err = UnpackEvent(eventSpec, topics, data, out)
	require.NoError(t, err)
	assert.Equal(t, in, out)
}

func splatPtr(v interface{}) []interface{} {
	rv := reflect.ValueOf(v).Elem()

	vals := make([]interface{}, rv.NumField())
	for i := 0; i < rv.NumField(); i++ {
		vals[i] = rv.Field(i).Addr().Interface()
	}

	return vals
}

func splat(v interface{}) []interface{} {
	rv := reflect.ValueOf(v).Elem()

	vals := make([]interface{}, rv.NumField())
	for i := 0; i < rv.NumField(); i++ {
		vals[i] = rv.Field(i).Interface()
	}

	return vals
}
