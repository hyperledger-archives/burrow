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
	"encoding/json"
	"runtime/debug"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var chainID = "myChainID"

var privateAccounts = make(map[crypto.Address]acm.AddressableSigner)

func makePrivateAccount(str string) *acm.PrivateAccount {
	acc := acm.GeneratePrivateAccountFromSecret(str)
	privateAccounts[acc.GetAddress()] = acc
	return acc
}

func TestSendTx(t *testing.T) {
	sendTx := &payload.SendTx{
		Inputs: []*payload.TxInput{
			{
				Address:  makePrivateAccount("input1").GetAddress(),
				Amount:   12345,
				Sequence: 67890,
			},
			{
				Address:  makePrivateAccount("input2").GetAddress(),
				Amount:   111,
				Sequence: 222,
			},
		},
		Outputs: []*payload.TxOutput{
			{
				Address: makePrivateAccount("output1").GetAddress(),
				Amount:  333,
			},
			{
				Address: makePrivateAccount("output2").GetAddress(),
				Amount:  444,
			},
		},
	}
	testTxMarshalJSON(t, sendTx)
	testTxSignVerify(t, sendTx)

	tx := Enclose("Foo", sendTx).Tx
	value, ok := tx.Tagged().Get("Inputs")
	require.True(t, ok)
	assert.Equal(t, sendTx.Inputs, value)

	value, ok = tx.Tagged().Get("ChainID")
	require.True(t, ok)
	assert.Equal(t, "Foo", value)
}

func TestCallTxSignable(t *testing.T) {
	toAddress := makePrivateAccount("contract1").GetAddress()
	callTx := &payload.CallTx{
		Input: &payload.TxInput{
			Address:  makePrivateAccount("input1").GetAddress(),
			Amount:   12345,
			Sequence: 67890,
		},
		Address:  &toAddress,
		GasLimit: 111,
		Fee:      222,
		Data:     []byte("data1"),
	}
	testTxMarshalJSON(t, callTx)
	testTxSignVerify(t, callTx)
}

func TestNameTxSignable(t *testing.T) {
	nameTx := &payload.NameTx{
		Input: &payload.TxInput{
			Address:  makePrivateAccount("input1").GetAddress(),
			Amount:   12345,
			Sequence: 250,
		},
		Name: "google.com",
		Data: "secretly.not.google.com",
		Fee:  1000,
	}
	testTxMarshalJSON(t, nameTx)
	testTxSignVerify(t, nameTx)
}

func TestBondTxSignable(t *testing.T) {
	bondTx := &payload.BondTx{
		Input: &payload.TxInput{
			Address:  makePrivateAccount("input1").GetAddress(),
			Amount:   12345,
			Sequence: 67890,
		},
	}
	testTxMarshalJSON(t, bondTx)
	testTxSignVerify(t, bondTx)
}

func TestUnbondTxSignable(t *testing.T) {
	unbondTx := &payload.UnbondTx{
		Input: &payload.TxInput{
			Address:  makePrivateAccount("input1").GetAddress(),
			Amount:   12345,
			Sequence: 67890,
		},
	}
	testTxMarshalJSON(t, unbondTx)
	testTxSignVerify(t, unbondTx)
}

func TestPermissionsTxSignable(t *testing.T) {
	permsTx := &payload.PermsTx{
		Input: &payload.TxInput{
			Address:  makePrivateAccount("input1").GetAddress(),
			Amount:   12345,
			Sequence: 250,
		},
		PermArgs: permission.SetBaseArgs(makePrivateAccount("address1").GetAddress(), 1, true),
	}

	testTxMarshalJSON(t, permsTx)
	testTxSignVerify(t, permsTx)
}

func TestTxWrapper_MarshalJSON(t *testing.T) {
	toAddress := makePrivateAccount("contract1").GetAddress()
	callTx := &payload.CallTx{
		Input: &payload.TxInput{
			Address:  makePrivateAccount("input1").GetAddress(),
			Amount:   12345,
			Sequence: 67890,
		},
		Address:  &toAddress,
		GasLimit: 111,
		Fee:      222,
		Data:     []byte("data1"),
	}
	testTxMarshalJSON(t, callTx)
	testTxSignVerify(t, callTx)

	tx := Enclose("Foo", callTx).Tx
	value, ok := tx.Tagged().Get("Input")
	require.True(t, ok)
	assert.Equal(t, callTx.Input, value)
}

func TestNewPermissionsTxWithSequence(t *testing.T) {
	privateAccount := makePrivateAccount("shhhhh")
	args := permission.SetBaseArgs(privateAccount.GetPublicKey().GetAddress(), permission.HasRole, true)
	permTx := payload.NewPermsTxWithSequence(privateAccount.GetPublicKey(), args, 1)
	testTxMarshalJSON(t, permTx)
	testTxSignVerify(t, permTx)
}

func testTxMarshalJSON(t *testing.T, tx payload.Payload) {
	txw := &Tx{Payload: tx}
	bs, err := json.Marshal(txw)
	require.NoError(t, err)
	txwOut := new(Tx)
	err = json.Unmarshal(bs, txwOut)
	require.NoError(t, err)
	bsOut, err := json.Marshal(txwOut)
	require.NoError(t, err)
	assert.Equal(t, string(bs), string(bsOut))
}

func testTxSignVerify(t *testing.T, tx payload.Payload) {
	inputs := tx.GetInputs()
	var signers []acm.AddressableSigner
	for _, in := range inputs {
		signers = append(signers, privateAccounts[in.Address])
	}
	txEnv := Enclose(chainID, tx)
	require.NoError(t, txEnv.Sign(signers...), "Error signing tx: %s", debug.Stack())
	require.NoError(t, txEnv.Verify(nil, chainID), "Error verifying tx: %s", debug.Stack())
}
