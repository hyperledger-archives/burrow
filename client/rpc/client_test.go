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

package rpc

import (
	"fmt"
	"testing"

	// "github.com/stretchr/testify/assert"

	mockclient "github.com/monax/eris-db/client/mock"
	mockkeys "github.com/monax/eris-db/keys/mock"
)

func Test(t *testing.T) {
	mockKeyClient := mockkeys.NewMockKeyClient()
	mockNodeClient := mockclient.NewMockNodeClient()
	testSend(t, mockNodeClient, mockKeyClient)
	testCall(t, mockNodeClient, mockKeyClient)
	testName(t, mockNodeClient, mockKeyClient)
	testPermissions(t, mockNodeClient, mockKeyClient)
	// t.Run("BondTransaction", )
	// t.Run("UnbondTransaction", )
	// t.Run("RebondTransaction", )
}

func testSend(t *testing.T,
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
	// assert.NotEqual(t, txSend)
	// TODO: test content of Transaction
}

func testCall(t *testing.T,
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

func testName(t *testing.T,
	nodeClient *mockclient.MockNodeClient, keyClient *mockkeys.MockKeyClient) {

	// generate an ED25519 key and ripemd160 address
	addressString := fmt.Sprintf("%X", keyClient.NewKey())
	// Public key can be queried from mockKeyClient.PublicKey(address)
	// but here we let the transaction factory retrieve the public key
	// which will then also overwrite the address we provide the function.
	// As a result we will assert whether address generated above, is identical
	// to address in generated transation.
	publicKeyString := ""
	// set an amount to transfer
	amountString := "1000"
	// unset nonce so that we retrieve nonce from account
	nonceString := ""
	// set fee
	feeString := "100"
	// set data
	dataString := fmt.Sprintf("%X", "We are DOUG.")
	// set name
	nameString := fmt.Sprintf("%s", "DOUG")

	_, err := Name(nodeClient, keyClient, publicKeyString, addressString,
		amountString, nonceString, feeString, nameString, dataString)
	if err != nil {
		t.Logf("Error in NameTx: %s", err)
		t.Fail()
	}
	// TODO: test content of Transaction
}

func testPermissions(t *testing.T,
	nodeClient *mockclient.MockNodeClient, keyClient *mockkeys.MockKeyClient) {

	// generate an ED25519 key and ripemd160 address
	addressString := fmt.Sprintf("%X", keyClient.NewKey())
	// Public key can be queried from mockKeyClient.PublicKey(address)
	// but here we let the transaction factory retrieve the public key
	// which will then also overwrite the address we provide the function.
	// As a result we will assert whether address generated above, is identical
	// to address in generated transation.
	publicKeyString := ""
	// generate an additional address to set permissions for
	permAddressString := fmt.Sprintf("%X", keyClient.NewKey())
	// unset nonce so that we retrieve nonce from account
	nonceString := ""

	_, err := Permissions(nodeClient, keyClient, publicKeyString, addressString,
		nonceString, "setBase", []string{permAddressString, "root", "true"})
	if err != nil {
		t.Logf("Error in PermissionsTx: %s", err)
		t.Fail()
	}
	// TODO: test content of Transaction
}
