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

package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/go-wire"
)

func TestPrivateAccountSerialise(t *testing.T) {
	type PrivateAccountContainingStruct struct {
		PrivateAccount PrivateAccount
		ChainID        string
	}
	// This test is really testing this go wire declaration in private_account.go
	//var _ = wire.RegisterInterface(struct{ PrivateAccount }{}, wire.ConcreteType{concretePrivateAccountWrapper{}, 0x01})

	acc := GeneratePrivateAccountFromSecret("Super Secret Secret")

	// Go wire cannot serialise a top-level interface type it needs to be a field or sub-field of a struct
	// at some depth. i.e. you MUST wrap an interface if you want it to be decoded (they do not document this)
	var accStruct = PrivateAccountContainingStruct{
		PrivateAccount: acc,
		ChainID:        "TestChain",
	}

	// We will write into this
	accStructOut := PrivateAccountContainingStruct{}

	// We must pass in a value type to read from (accStruct), but provide a pointer type to write into (accStructout
	wire.ReadBinaryBytes(wire.BinaryBytes(accStruct), &accStructOut)

	assert.Equal(t, accStruct, accStructOut)
}
