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

package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePrivateKey(t *testing.T) {
	privateKey, err := GeneratePrivateKey(bytes.NewBuffer([]byte{
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
		1, 2, 3, 4, 5, 6, 7, 8,
	}), CurveTypeEd25519)
	require.NoError(t, err)
	assert.NoError(t, EnsureEd25519PrivateKeyCorrect(privateKey.RawBytes()))
	badKey := privateKey.RawBytes()
	// Change part of the public part to not match private part
	badKey[35] = 2
	assert.Error(t, EnsureEd25519PrivateKeyCorrect(badKey))
	goodKey := privateKey.RawBytes()
	// Change part of the private part invalidating public part
	goodKey[31] = 2
	assert.Error(t, EnsureEd25519PrivateKeyCorrect(badKey))
}
