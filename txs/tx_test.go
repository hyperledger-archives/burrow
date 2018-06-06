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

package txs

import (
	"fmt"
	"testing"

	"encoding/json"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	ptypes "github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/permission/snatives"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var chainID = "myChainID"

func makeAddress(str string) (address crypto.Address) {
	copy(address[:], ([]byte)(str))
	return
}

func TestSendTxSignable(t *testing.T) {
	sendTx := &SendTx{
		Inputs: []*TxInput{
			{
				Address:  makeAddress("input1"),
				Amount:   12345,
				Sequence: 67890,
			},
			{
				Address:  makeAddress("input2"),
				Amount:   111,
				Sequence: 222,
			},
		},
		Outputs: []*TxOutput{
			{
				Address: makeAddress("output1"),
				Amount:  333,
			},
			{
				Address: makeAddress("output2"),
				Amount:  444,
			},
		},
	}
	signBytes := ChainWrap(chainID, sendTx).MustSignBytes()
	signStr := string(signBytes)
	expected := fmt.Sprintf(`{"chain_id":"%s","tx":[1,{"inputs":[{"address":"%s","amount":12345,"sequence":67890},{"address":"%s","amount":111,"sequence":222}],"outputs":[{"address":"%s","amount":333},{"address":"%s","amount":444}]}]}`,
		chainID, sendTx.Inputs[0].Address.String(), sendTx.Inputs[1].Address.String(), sendTx.Outputs[0].Address.String(), sendTx.Outputs[1].Address.String())

	if signStr != expected {
		t.Errorf("Got unexpected sign string for SendTx. Expected:\n%v\nGot:\n%v", expected, signStr)
	}
}

func TestCallTxSignable(t *testing.T) {
	toAddress := makeAddress("contract1")
	callTx := &CallTx{
		Input: &TxInput{
			Address:  makeAddress("input1"),
			Amount:   12345,
			Sequence: 67890,
		},
		Address:  &toAddress,
		GasLimit: 111,
		Fee:      222,
		Data:     []byte("data1"),
	}
	signBytes := ChainWrap(chainID, callTx).MustSignBytes()
	signStr := string(signBytes)
	expected := fmt.Sprintf(`{"chain_id":"%s","tx":[2,{"address":"%s","data":"6461746131","fee":222,"gas_limit":111,"input":{"address":"%s","amount":12345,"sequence":67890}}]}`,
		chainID, callTx.Address.String(), callTx.Input.Address.String())
	if signStr != expected {
		t.Errorf("Got unexpected sign string for CallTx. Expected:\n%v\nGot:\n%v", expected, signStr)
	}
}

func TestNameTxSignable(t *testing.T) {
	nameTx := &NameTx{
		Input: &TxInput{
			Address:  makeAddress("input1"),
			Amount:   12345,
			Sequence: 250,
		},
		Name: "google.com",
		Data: "secretly.not.google.com",
		Fee:  1000,
	}
	signBytes := ChainWrap(chainID, nameTx).MustSignBytes()
	signStr := string(signBytes)
	expected := fmt.Sprintf(`{"chain_id":"%s","tx":[3,{"data":"secretly.not.google.com","fee":1000,"input":{"address":"%s","amount":12345,"sequence":250},"name":"google.com"}]}`,
		chainID, nameTx.Input.Address.String())
	if signStr != expected {
		t.Errorf("Got unexpected sign string for CallTx. Expected:\n%v\nGot:\n%v", expected, signStr)
	}
}

func TestBondTxSignable(t *testing.T) {
	privAccount := acm.GeneratePrivateAccountFromSecret("foooobars")
	bondTx := &BondTx{
		PubKey: privAccount.PublicKey(),
		Inputs: []*TxInput{
			{
				Address:  makeAddress("input1"),
				Amount:   12345,
				Sequence: 67890,
			},
			{
				Address:  makeAddress("input2"),
				Amount:   111,
				Sequence: 222,
			},
		},
		UnbondTo: []*TxOutput{
			{
				Address: makeAddress("output1"),
				Amount:  333,
			},
			{
				Address: makeAddress("output2"),
				Amount:  444,
			},
		},
	}
	expected := fmt.Sprintf(`{"chain_id":"%s",`+
		`"tx":[17,{"inputs":[{"address":"%s",`+
		`"amount":12345,"sequence":67890},{"address":"%s",`+
		`"amount":111,"sequence":222}],"pub_key":{"CurveType":1,"PublicKey":"%X"},`+
		`"unbond_to":[{"address":"%s",`+
		`"amount":333},{"address":"%s",`+
		`"amount":444}]}]}`,
		chainID,
		bondTx.Inputs[0].Address.String(),
		bondTx.Inputs[1].Address.String(),
		bondTx.PubKey.RawBytes(),
		bondTx.UnbondTo[0].Address.String(),
		bondTx.UnbondTo[1].Address.String())

	signBytes := ChainWrap(chainID, bondTx).MustSignBytes()
	assert.Equal(t, expected, string(signBytes), "Unexpected sign string for BondTx")
}

func TestUnbondTxSignable(t *testing.T) {
	unbondTx := &UnbondTx{
		Address: makeAddress("address1"),
		Height:  111,
	}
	signBytes := ChainWrap(chainID, unbondTx).MustSignBytes()
	signStr := string(signBytes)
	expected := fmt.Sprintf(`{"chain_id":"%s","tx":[18,{"address":"%s","height":111}]}`,
		chainID, unbondTx.Address.String())
	if signStr != expected {
		t.Errorf("Got unexpected sign string for UnbondTx")
	}
}

func TestPermissionsTxSignable(t *testing.T) {
	permsTx := &PermissionsTx{
		Input: &TxInput{
			Address:  makeAddress("input1"),
			Amount:   12345,
			Sequence: 250,
		},
		PermArgs: snatives.SetBaseArgs(makeAddress("address1"), 1, true),
	}

	body := ChainWrap(chainID, permsTx)
	env := Envelope{
		Body: *body,
	}
	signBytes := body.MustSignBytes()
	bodyOut := new(Body)

	bs, err := json.MarshalIndent(env, "", "  ")
	require.NoError(t, err)
	fmt.Println(string(bs))
	require.NoError(t, json.Unmarshal(signBytes, bodyOut))
	assert.Equal(t, signBytes, body.MustSignBytes())
}

func TestTxWrapper_MarshalJSON(t *testing.T) {
	toAddress := makeAddress("contract1")
	callTx := &CallTx{
		Input: &TxInput{
			Address:  makeAddress("input1"),
			Amount:   12345,
			Sequence: 67890,
		},
		Address:  &toAddress,
		GasLimit: 111,
		Fee:      222,
		Data:     []byte("data1"),
	}
	testTxMarshalJSON(t, callTx)
}

func TestNewPermissionsTxWithSequence(t *testing.T) {
	privateKey := crypto.PrivateKeyFromSecret("Shhh...", crypto.CurveTypeEd25519)

	args := snatives.SetBaseArgs(privateKey.GetPublicKey().Address(), ptypes.HasRole, true)
	permTx := NewPermissionsTxWithSequence(privateKey.GetPublicKey(), args, 1)
	testTxMarshalJSON(t, permTx)
}

func testTxMarshalJSON(t *testing.T, tx Tx) {
	txw := &Body{Tx: tx}
	bs, err := json.Marshal(txw)
	require.NoError(t, err)
	txwOut := new(Body)
	err = json.Unmarshal(bs, txwOut)
	require.NoError(t, err)
	bsOut, err := json.Marshal(txwOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}
