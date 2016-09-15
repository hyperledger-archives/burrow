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

func TestTransactionFactory(t *testing.T) {
	mockKeyClient := mockkeys.NewMockKeyClient()
	mockNodeClient := mockclient.NewMockNodeClient()
	testTransactionFactorySend(t, mockNodeClient, mockKeyClient)
	// t.Run("NameTransaction", )
	testTransactionFactoryCall(t, mockNodeClient, mockKeyClient)
	// t.Run("PermissionTransaction", )
	// t.Run("BondTransaction", )
	// t.Run("UnbondTransaction", )
	// t.Run("RebondTransaction", )
}

func testTransactionFactorySend(t *testing.T,
	nodeClient *mockclient.MockNodeClient, keyClient *mockkeys.MockKeyClient) {

	// generate an ED25519 key and ripemd160 address
	addressString := fmt.Sprintf("%X", keyClient.NewKey())
	// Public key can be queried from mockKeyClient.PublicKey(address)
	// but here we let the transaction factory retrieve the public key
	// which will then also overwrite the address we provide the function.
	// As a result we will assert whether address generated above, is identical
	// to address in generated transation.
	publicKeyString := ""
	// generate an additional address to send amount to
	toAddressString := fmt.Sprintf("%X", keyClient.NewKey())
	// set an amount to transfer
	amountString := "1000"
	// unset nonce so that we retrieve nonce from account
	nonceString := ""

	_, err := Send(nodeClient, keyClient, publicKeyString, addressString,
		toAddressString, amountString, nonceString)
	if err != nil {
		t.Logf("Error in SendTx: %s", err)
		t.Fail()
	}
	// TODO: test content of Transaction
}

func testTransactionFactoryCall(t *testing.T,
	nodeClient *mockclient.MockNodeClient, keyClient *mockkeys.MockKeyClient) {

	// generate an ED25519 key and ripemd160 address
	addressString := fmt.Sprintf("%X", keyClient.NewKey())
	// Public key can be queried from mockKeyClient.PublicKey(address)
	// but here we let the transaction factory retrieve the public key
	// which will then also overwrite the address we provide the function.
	// As a result we will assert whether address generated above, is identical
	// to address in generated transation.
	publicKeyString := ""
	// generate an additional address to send amount to
	toAddressString := fmt.Sprintf("%X", keyClient.NewKey())
	// set an amount to transfer
	amountString := "1000"
	// unset nonce so that we retrieve nonce from account
	nonceString := ""
	// set gas 
	gasString := "1000"
	// set fee
	feeString := "100"
	// set data
	dataString := fmt.Sprintf("%X", "We are DOUG.")

	_, err := Call(nodeClient, keyClient, publicKeyString, addressString,
		toAddressString, amountString, nonceString, gasString, feeString, dataString)
	if err != nil {
		t.Logf("Error in CallTx: %s", err)
		t.Fail()
	}
	// TODO: test content of Transaction
}