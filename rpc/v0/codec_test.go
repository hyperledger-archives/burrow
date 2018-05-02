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

package v0

import (
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeysEncoding(t *testing.T) {
	codec := NewTCodec()
	privateKey := crypto.PrivateKeyFromSecret("foo", crypto.CurveTypeEd25519)
	type keyPair struct {
		PrivateKey crypto.PrivateKey
		PublicKey  crypto.PublicKey
	}

	kp := keyPair{
		PrivateKey: privateKey,
		PublicKey:  privateKey.GetPublicKey(),
	}

	bs, err := codec.EncodeBytes(kp)
	require.NoError(t, err)

	kpOut := keyPair{}
	codec.DecodeBytes(&kpOut, bs)

	assert.Equal(t, kp, kpOut)
}
