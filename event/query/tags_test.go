package query

import (
	"reflect"
	"testing"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetReflect(t *testing.T) {
	type testTaggable struct {
		Foo     string
		Bar     string
		Baz     binary.HexBytes
		Address crypto.Address
		Indices []int
	}

	t.Run("Basic", func(t *testing.T) {
		tt := &testTaggable{
			Foo:     "Thumbs",
			Bar:     "Numbed",
			Baz:     []byte{255, 255, 255},
			Address: crypto.Address{1, 2, 3},
			Indices: []int{5, 7, 9},
		}

		rv := reflect.ValueOf(tt)
		value, ok := GetReflect(rv, "Foo")
		assert.True(t, ok)
		assert.Equal(t, tt.Foo, value)

		value, ok = GetReflect(rv, "Bar")
		assert.True(t, ok)
		assert.Equal(t, tt.Bar, value)

		value, ok = GetReflect(rv, "Baz")
		assert.True(t, ok)
		assert.Equal(t, binary.HexBytes{0xFF, 0xFF, 0xFF}, value)

		value, ok = GetReflect(rv, "Indices")
		assert.True(t, ok)
		assert.Equal(t, []int{5, 7, 9}, value)

		value, ok = GetReflect(rv, "Address")
		assert.True(t, ok)
		assert.Equal(t, crypto.MustAddressFromHexString("0102030000000000000000000000000000000000"), value)

		// Make sure we see updates through pointer
		tt.Foo = "Plums"
		value, ok = GetReflect(rv, "Foo")
		assert.True(t, ok)
		assert.Equal(t, tt.Foo, value)
	})

	type recursiveTaggable struct {
		Tags1 testTaggable
		Tags2 *testTaggable
		Self  *recursiveTaggable
	}
	t.Run("Recursive", func(t *testing.T) {
		tt := &recursiveTaggable{
			Tags1: testTaggable{
				Foo:     "Thumbs",
				Bar:     "Numbed",
				Baz:     []byte{255, 255, 255},
				Address: crypto.Address{1, 2, 3},
				Indices: []int{5, 7, 9},
			},
			Tags2: &testTaggable{
				Foo:     "Thumbs2",
				Bar:     "Numbed2",
				Baz:     []byte{255, 255, 255},
				Address: crypto.Address{1, 2, 3},
				Indices: []int{4, 3, 2},
			},
			Self: &recursiveTaggable{
				Tags1: testTaggable{
					Foo: "DeepFoo",
				},
				Self: &recursiveTaggable{
					Tags2: &testTaggable{
						Bar: "ReallyDeepBar",
					},
				},
			},
		}
		rv := reflect.ValueOf(tt)
		v, ok := GetReflect(rv, "Tags1.Foo")
		require.True(t, ok)
		require.Equal(t, "Thumbs", v)

		// Shouldn't get this deep by default
		v, ok = GetReflectDepth(rv, "Self.Tags1.Foo", 1)
		require.False(t, ok)

		v, ok = GetReflectDepth(rv, "Self.Tags1.Foo", 2)
		require.True(t, ok)
		require.Equal(t, "DeepFoo", v)

		v, ok = GetReflectDepth(rv, "Self.Self.Tags2.Bar", 3)
		require.True(t, ok)
		require.Equal(t, "ReallyDeepBar", v)
	})

}

func TestReflectTagged_nil(t *testing.T) {
	type testStruct struct {
		Foo string
	}

	var ts *testStruct

	value, ok := GetReflect(reflect.ValueOf(ts), "Foo")
	assert.False(t, ok)
	assert.Nil(t, value)
}
