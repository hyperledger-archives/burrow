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
	"encoding/json"
	"fmt"
	"testing"

	// "github.com/stretchr/testify/assert"

	mockclient "github.com/hyperledger/burrow/client/mock"
	mockkeys "github.com/hyperledger/burrow/keys/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	mockKeyClient := mockkeys.NewKeyClient()
	mockNodeClient := mockclient.NewMockNodeClient()
	testSend(t, mockNodeClient, mockKeyClient)
}

func TestCall(t *testing.T) {
	mockKeyClient := mockkeys.NewKeyClient()
	mockNodeClient := mockclient.NewMockNodeClient()
	testCall(t, mockNodeClient, mockKeyClient)
}

func TestName(t *testing.T) {
	mockKeyClient := mockkeys.NewKeyClient()
	mockNodeClient := mockclient.NewMockNodeClient()
	testName(t, mockNodeClient, mockKeyClient)
}

func TestPermissions(t *testing.T) {
	mockKeyClient := mockkeys.NewKeyClient()
	mockNodeClient := mockclient.NewMockNodeClient()
	testPermissions(t, mockNodeClient, mockKeyClient)
}

func testSend(t *testing.T,
	nodeClient *mockclient.MockNodeClient, keyClient *mockkeys.KeyClient) {

	// generate an ED25519 key and ripemd160 address
	addressString := keyClient.NewKey("").String()
	// Public key can be queried from mockKeyClient.PublicKey(address)
	// but here we let the transaction factory retrieve the public key
	// which will then also overwrite the address we provide the function.
	// As a result we will assert whether address generated above, is identical
	// to address in generated transation.
	publicKeyString := ""
	// generate an additional address to send amount to
	toAddressString := keyClient.NewKey("").String()
	// set an amount to transfer
	amountString := "1000"
	// unset sequence so that we retrieve sequence from account
	sequenceString := ""

	_, err := Send(nodeClient, keyClient, publicKeyString, addressString,
		toAddressString, amountString, sequenceString)
	require.NoError(t, err, "Error in Send")
	// assert.NotEqual(t, txSend)
	// TODO: test content of Transaction
}

func testCall(t *testing.T,
	nodeClient *mockclient.MockNodeClient, keyClient *mockkeys.KeyClient) {

	// generate an ED25519 key and ripemd160 address
	addressString := keyClient.NewKey("").String()
	// Public key can be queried from mockKeyClient.PublicKey(address)
	// but here we let the transaction factory retrieve the public key
	// which will then also overwrite the address we provide the function.
	// As a result we will assert whether address generated above, is identical
	// to address in generated transation.
	publicKeyString := ""
	// generate an additional address to send amount to
	toAddressString := keyClient.NewKey("").String()
	// set an amount to transfer
	amountString := "1000"
	// unset sequence so that we retrieve sequence from account
	sequenceString := ""
	// set gas
	gasString := "1000"
	// set fee
	feeString := "100"
	// set data
	dataString := fmt.Sprintf("%X", "We are DOUG.")

	_, err := Call(nodeClient, keyClient, publicKeyString, addressString,
		toAddressString, amountString, sequenceString, gasString, feeString, dataString)
	if err != nil {
		t.Logf("Error in CallTx: %s", err)
		t.Fail()
	}
	// TODO: test content of Transaction
}

func testName(t *testing.T,
	nodeClient *mockclient.MockNodeClient, keyClient *mockkeys.KeyClient) {

	// generate an ED25519 key and ripemd160 address
	addressString := keyClient.NewKey("").String()
	// Public key can be queried from mockKeyClient.PublicKey(address)
	// but here we let the transaction factory retrieve the public key
	// which will then also overwrite the address we provide the function.
	// As a result we will assert whether address generated above, is identical
	// to address in generated transation.
	publicKeyString := ""
	// set an amount to transfer
	amountString := "1000"
	// unset sequence so that we retrieve sequence from account
	sequenceString := ""
	// set fee
	feeString := "100"
	// set data
	dataString := fmt.Sprintf("%X", "We are DOUG.")
	// set name
	nameString := "DOUG"

	_, err := Name(nodeClient, keyClient, publicKeyString, addressString,
		amountString, sequenceString, feeString, nameString, dataString)
	if err != nil {
		t.Logf("Error in NameTx: %s", err)
		t.Fail()
	}
	// TODO: test content of Transaction
}

func testPermissions(t *testing.T,
	nodeClient *mockclient.MockNodeClient, keyClient *mockkeys.KeyClient) {

	// generate an ED25519 key and ripemd160 address
	addressString := keyClient.NewKey("").String()
	// Public key can be queried from mockKeyClient.PublicKey(address)
	// but here we let the transaction factory retrieve the public key
	// which will then also overwrite the address we provide the function.
	// As a result we will assert whether address generated above, is identical
	// to address in generated transation.
	publicKeyString := ""
	// generate an additional address to set permissions for
	permAddressString := keyClient.NewKey("").String()
	// unset sequence so that we retrieve sequence from account
	sequenceString := ""

	tx, err := Permissions(nodeClient, keyClient, publicKeyString, addressString,
		sequenceString, "setBase", permAddressString, "root", "", "true")
	if err != nil {
		t.Logf("Error in PermissionsTx: %s", err)
		t.Fail()
	}

	bs, err := json.Marshal(tx.PermArgs)
	require.NoError(t, err)
	expected := fmt.Sprintf(`{"PermFlag":256,"Address":"%s","Permission":1,"Value":true}`, permAddressString)
	assert.Equal(t, expected, string(bs))
}
