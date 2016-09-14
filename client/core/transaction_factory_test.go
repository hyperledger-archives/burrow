// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"testing"

	"github.com/eris-ltd/eris-keys/crypto"

	"github.com/eris-ltd/eris-db/client"
	"github.com/eris-ltd/eris-db/keys"
)

// Unit tests for client/core an idael showcase for the need to
// modularise and restructure the components of the code.


func TestTransactionFactory(t *testing.T) {
	// test in parallel
	t.Run("ExtractInputAddress from transaction", func (t *testing.T) {
		t.Run("SendTransaction", testTransactionFactorySend)
		// t.Run("NameTransaction", )
		// t.Run("CallTransaction", )
		// t.Run("PermissionTransaction", )
		// t.Run("BondTransaction", )
		// t.Run("UnbondTransaction", )
		// t.Run("RebondTransaction", )
	})
}

func testTransactionFactorySend(t *testing.T) {
	mockKeyClient := keys.NewMockKeyClient()
	mockNodeClient := client.NewMockNodeClient()

	// generate an ED25519 key and ripemd160 address
	address := mockKeyClient.NewKey()
	// Public key can be queried from mockKeyClient.PublicKey(address)
	// but here we let the transaction factory retrieve the public key
	// which will then also overwrite the address we provide the function.
	// As a result we will assert whether address generated above, is identical
	// to address in generated transation.
	publicKey := ""
	// generate an additional address to send amount to
	toAddress := mockKeyClient.NewKey()
	// set an amount to transfer
	amount := "1000"
	// set nonce to unset
	nonce := ""

	tx, err := Send(mockNodeClient, mockKeyClient, publicKey, string(address),
		string(toAddress), amount, nonce)
	if err != nil {
		t.Fail()
	}
}

