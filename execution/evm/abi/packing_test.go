package abi

import (
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackEvent(t *testing.T) {
	t.Run("simple event", func(t *testing.T) {
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
	})

	t.Run("EventEmitter", func(t *testing.T) {
		type args struct {
			Direction []byte
			Trueism   bool
			German    string
			NewDepth  *big.Int
			Bignum    int8
			Hash      string
		}
		spec, err := ReadSpec(solidity.Abi_EventEmitter)
		require.NoError(t, err)

		eventSpec := spec.EventsByName["ManyTypes"]

		dir := make([]byte, 32)
		copy(dir, "frogs")
		bignum := big.NewInt(1000)
		in := args{
			Direction: dir,
			Trueism:   false,
			German:    "foo",
			NewDepth:  bignum,
			Bignum:    100,
			Hash:      "ba",
		}
		topics, data, err := PackEvent(&eventSpec, in)
		require.NoError(t, err)

		out := new(args)
		err = UnpackEvent(&eventSpec, topics, data, out)
		require.NoError(t, err)
	})

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
