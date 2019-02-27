package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
	err := EnsureKeyFormatStore(keyStore)
	require.NoError(t, err)

	err = EnsureKeyFormatStore(&keyStore)
	require.NoError(t, err, "pointer to keystore should work")

	keyStore = testKeyStore{
		Accounts: NewMustKeyFormat("foo", 4, 5, 6),
		Storage:  NewMustKeyFormat("foo", 4, 5, 6),
	}
	err = EnsureKeyFormatStore(&keyStore)
	require.Error(t, err, "duplicate prefixes should be detected")

	// Test missing formats
	keyStore = testKeyStore{}
	err = EnsureKeyFormatStore(&keyStore)
	require.Error(t, err, "all formats should be set")

	keyStore = testKeyStore{
		Accounts: NewMustKeyFormat("foo", 4, 5, 6),
	}
	err = EnsureKeyFormatStore(&keyStore)
	require.Error(t, err, "all formats should be set")

	keyStore2 := struct {
		Accounts MustKeyFormat
		Storage  *MustKeyFormat
	}{
		Accounts: *NewMustKeyFormat("foo", 56, 6),
		Storage:  NewMustKeyFormat("foo2", 1, 2),
	}

	err = EnsureKeyFormatStore(keyStore2)
	require.NoError(t, err)

	keyStore2 = struct {
		Accounts MustKeyFormat
		Storage  *MustKeyFormat
	}{
		Storage: NewMustKeyFormat("foo2", 1, 2),
	}
	err = EnsureKeyFormatStore(keyStore2)
	require.NoError(t, err)

	err = EnsureKeyFormatStore(keyStore2)
	require.NoError(t, err)

	keyStore2 = struct {
		Accounts MustKeyFormat
		Storage  *MustKeyFormat
	}{
		Accounts: *NewMustKeyFormat("foo", 56, 6),
		Storage:  NewMustKeyFormat("foo", 1, 2),
	}

	err = EnsureKeyFormatStore(keyStore2)
	require.Error(t, err, "duplicate prefixes should be detected")
}
