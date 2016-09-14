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
	"fmt"
	"testing"

	mockclient "github.com/eris-ltd/eris-db/client/mock"
	mockkeys "github.com/eris-ltd/eris-db/keys/mock"
)

// func TestTransactionFactory(t *testing.T) {
// 	// test in parallel
// 	t.Run("ExtractInputAddress from transaction", func (t1 *testing.T) {
// 		t1.Run("SendTransaction", testTransactionFactorySend)
// 		// t.Run("NameTransaction", )
// 		// t.Run("CallTransaction", )
// 		// t.Run("PermissionTransaction", )
// 		// t.Run("BondTransaction", )
// 		// t.Run("UnbondTransaction", )
// 		// t.Run("RebondTransaction", )
// 	})
// }

func TestTransactionFactorySend(t *testing.T) {
	mockKeyClient := mockkeys.NewMockKeyClient()
	mockNodeClient := mockclient.NewMockNodeClient()

	// generate an ED25519 key and ripemd160 address
	addressString := fmt.Sprintf("%X", mockKeyClient.NewKey())
	// Public key can be queried from mockKeyClient.PublicKey(address)
	// but here we let the transaction factory retrieve the public key
	// which will then also overwrite the address we provide the function.
	// As a result we will assert whether address generated above, is identical
	// to address in generated transation.
	publicKeyString := ""
	// generate an additional address to send amount to
	toAddressString := fmt.Sprintf("%X", mockKeyClient.NewKey())
	// set an amount to transfer
	amount := "1000"
	// unset nonce so that we retrieve nonce from account
	nonce := ""

	_, err := Send(mockNodeClient, mockKeyClient, publicKeyString, addressString,
		toAddressString, amount, nonce)
	if err != nil {
		t.Logf("Error: %s", err)
		t.Fail()
	}
}
