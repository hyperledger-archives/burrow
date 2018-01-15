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

	acm "github.com/hyperledger/burrow/account"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeysEncoding(t *testing.T) {
	codec := NewTCodec()
	privateKey := acm.PrivateKeyFromSecret("foo")
	keyPair := struct {
		PrivateKey acm.PrivateKey
		PublicKey  acm.PublicKey
	}{
		PrivateKey: privateKey,
		PublicKey:  privateKey.PublicKey(),
	}

	bs, err := codec.EncodeBytes(keyPair)
	require.NoError(t, err)
	assert.Equal(t, `{"PrivateKey":[1,"2C26B46B68FFC68FF99B453C1D30413413422D706483BFA0F98A5E886266E7AE34D26579DBB456693E540672CF922F52DDE0D6532E35BF06BE013A7C532F20E0"],"PublicKey":[1,"34D26579DBB456693E540672CF922F52DDE0D6532E35BF06BE013A7C532F20E0"]}`,
		string(bs))

}
