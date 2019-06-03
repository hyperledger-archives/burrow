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
