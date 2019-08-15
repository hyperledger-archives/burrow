package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyFormatBytes(t *testing.T) {
	kf := NewMustKeyFormat("e", 8, 8, 8)
	assert.Panics(t, func() { kf.KeyBytes([]byte{1, 2, 3}) }, "insufficient bytes for segment should panic")
	assert.Equal(t, []byte{'e', 1, 2, 3, 4, 5, 6, 7, 8}, kf.KeyBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8}))
	assert.Equal(t, []byte{'e', 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 1, 1, 2, 2, 3, 3},
		kf.KeyBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8}, []byte{1, 2, 3, 4, 5, 6, 7, 8}, []byte{0, 0, 1, 1, 2, 2, 3, 3}))
	assert.Equal(t, []byte{'e'}, kf.KeyBytes())
}

func TestKeyFormat(t *testing.T) {
	kf := NewMustKeyFormat("eab", 8, 8, 8)
	key := []byte{'e', 'a', 'b', 0, 0, 0, 0, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0, 200, 0, 0, 0, 0, 0, 0, 1, 144}
	var a, b, c int64 = 100, 200, 400
	assert.Equal(t, key, kf.Key(a, b, c))

	var ao, bo, co = new(int64), new(int64), new(int64)
	kf.Scan(key, ao, bo, co)
	assert.Equal(t, a, *ao)
	assert.Equal(t, b, *bo)
	assert.Equal(t, c, *co)

	// Allow for omitted values
	bo, co = new(int64), new(int64)
	kf.Scan(key, nil, bo, co)
	assert.Equal(t, b, *bo)
	assert.Equal(t, c, *co)

	ao, bo, _ = new(int64), new(int64), new(int64)
	bs := new([]byte)
	kf.Scan(key, ao, bo, bs)
	assert.Equal(t, a, *ao)
	assert.Equal(t, b, *bo)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 1, 144}, *bs)

	assert.Equal(t, []byte{'e', 'a', 'b', 0, 0, 0, 0, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0, 200}, kf.Key(a, b))
}

func TestNegativeKeys(t *testing.T) {
	kf := NewMustKeyFormat("e", 8, 8)

	var a, b int64 = -100, -200
	// One's complement plus one
	key := []byte{'e',
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(0xff + a + 1),
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, byte(0xff + b + 1)}
	assert.Equal(t, key, kf.Key(a, b))

	var ao, bo = new(int64), new(int64)
	kf.Scan(key, ao, bo)
	assert.Equal(t, a, *ao)
	assert.Equal(t, b, *bo)
}

func TestOverflow(t *testing.T) {
	kf := NewMustKeyFormat("o", 8, 8)

	var a int64 = 1 << 62
	var b uint64 = 1 << 63
	key := []byte{'o',
		0x40, 0, 0, 0, 0, 0, 0, 0,
		0x80, 0, 0, 0, 0, 0, 0, 0,
	}
	assert.Equal(t, key, kf.Key(a, b))

	var ao, bo = new(int64), new(int64)
	kf.Scan(key, ao, bo)
	assert.Equal(t, a, *ao)
	assert.Equal(t, int64(b), *bo)
}

func TestVariadicKeys(t *testing.T) {
	kf := NewMustKeyFormat("o", 8, 0)
	str := "hello this is an arbitrary length string"

	var a int64 = 1 << 62
	key := append([]byte{'o', 0x40, 0, 0, 0, 0, 0, 0, 0}, []byte(str)...)

	assert.Equal(t, key, kf.Key(a, str))

	var ao, stro = new(int64), new(string)
	kf.Scan(key, ao, stro)
	assert.Equal(t, a, *ao)
	assert.Equal(t, str, *stro)
}

func TestKeyFormat_Fix(t *testing.T) {
	kf := NewMustKeyFormat("o", 8, 8, 4)
	prefix := string([]byte{'o', 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 4})
	assert.Equal(t, *NewMustKeyFormat(prefix, 4), *kf.Fix(3, 4))
}

func TestKeyFormat_Suffix(t *testing.T) {
	kf := NewMustKeyFormat("diplodocus", 4, 0)
	key := kf.KeyNoPrefix([]byte("Hi, "), "dinosaur")
	assert.Equal(t, "Hi, dinosaur", key.String())
}

func TestKeyFormat_Layout(t *testing.T) {
	kf := NewMustKeyFormat("diplodocus", 4, DelimiterSegmentLength, VariadicSegmentLength)
	key := kf.KeyNoPrefix([]byte("Hi, "), "dinosaur")
	assert.Equal(t, "Hi, dinosaur", key.String())
}

type testKeyStore = struct {
	Accounts *MustKeyFormat
	Storage  *MustKeyFormat
	foo      string
}

func TestEnsureKeyStore(t *testing.T) {
	keyStore := testKeyStore{
		Accounts: NewMustKeyFormat("foo", 4, 5, 6),
		Storage:  NewMustKeyFormat("foos", 4, 5, 6),
	}
	_, err := EnsureKeyFormatStore(keyStore)
	require.NoError(t, err)

	_, err = EnsureKeyFormatStore(&keyStore)
	require.NoError(t, err, "pointer to keystore should work")

	keyStore = testKeyStore{
		Accounts: NewMustKeyFormat("foo", 4, 5, 6),
		Storage:  NewMustKeyFormat("foo", 4, 5, 6),
	}
	_, err = EnsureKeyFormatStore(&keyStore)
	require.Error(t, err, "duplicate prefixes should be detected")

	// Test missing formats
	keyStore = testKeyStore{}
	_, err = EnsureKeyFormatStore(&keyStore)
	require.Error(t, err, "all formats should be set")

	keyStore = testKeyStore{
		Accounts: NewMustKeyFormat("foo", 4, 5, 6),
	}
	_, err = EnsureKeyFormatStore(&keyStore)
	require.Error(t, err, "all formats should be set")

	keyStore2 := struct {
		Accounts MustKeyFormat
		Storage  *MustKeyFormat
	}{
		Accounts: *NewMustKeyFormat("foo", 56, 6),
		Storage:  NewMustKeyFormat("foo2", 1, 2),
	}

	_, err = EnsureKeyFormatStore(keyStore2)
	require.NoError(t, err)

	keyStore2 = struct {
		Accounts MustKeyFormat
		Storage  *MustKeyFormat
	}{
		Storage: NewMustKeyFormat("foo2", 1, 2),
	}
	_, err = EnsureKeyFormatStore(keyStore2)
	require.NoError(t, err)

	_, err = EnsureKeyFormatStore(keyStore2)
	require.NoError(t, err)

	keyStore2 = struct {
		Accounts MustKeyFormat
		Storage  *MustKeyFormat
	}{
		Accounts: *NewMustKeyFormat("foo", 56, 6),
		Storage:  NewMustKeyFormat("foo", 1, 2),
	}

	_, err = EnsureKeyFormatStore(keyStore2)
	require.Error(t, err, "duplicate prefixes should be detected")
}
