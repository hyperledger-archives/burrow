// Copyright 2019 Monax Industries Limited
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
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAminoEncodeTxDecodeTx(t *testing.T) {
	codec := NewAminoCodec()
	inputAddress := crypto.Address{1, 2, 3, 4, 5}
	outputAddress := crypto.Address{5, 4, 3, 2, 1}
	amount := uint64(2)
	sequence := uint64(3)
	tx := &payload.SendTx{
		Inputs: []*payload.TxInput{{
			Address:  inputAddress,
			Amount:   amount,
			Sequence: sequence,
		}},
		Outputs: []*payload.TxOutput{{
			Address: outputAddress,
			Amount:  amount,
		}},
	}
	txEnv := Enclose(chainID, tx)
	txBytes, err := codec.EncodeTx(txEnv)
	if err != nil {
		t.Fatal(err)
	}
	txEnvOut, err := codec.DecodeTx(txBytes)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, txEnv, txEnvOut)
}

func TestAminoEncodeTxDecodeTx_CallTx(t *testing.T) {
	codec := NewAminoCodec()
	inputAccount := acm.GeneratePrivateAccountFromSecret("fooo")
	amount := uint64(2)
	sequence := uint64(3)
	tx := &payload.CallTx{
		Input: &payload.TxInput{
			Address:  inputAccount.GetAddress(),
			Amount:   amount,
			Sequence: sequence,
		},
		GasLimit: 233,
		Fee:      2,
		Address:  nil,
		Data:     []byte("code"),
	}
	txEnv := Enclose(chainID, tx)
	require.NoError(t, txEnv.Sign(inputAccount))
	txBytes, err := codec.EncodeTx(txEnv)
	if err != nil {
		t.Fatal(err)
	}
	txEnvOut, err := codec.DecodeTx(txBytes)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, txEnv, txEnvOut)
}

func TestAminoTxEnvelope(t *testing.T) {
	codec := NewAminoCodec()
	privAccFrom := acm.GeneratePrivateAccountFromSecret("foo")
	privAccTo := acm.GeneratePrivateAccountFromSecret("bar")
	toAddress := privAccTo.GetAddress()
	txEnv := Enclose("testChain", payload.NewCallTxWithSequence(privAccFrom.GetPublicKey(), &toAddress,
		[]byte{3, 4, 5, 5}, 343, 2323, 12, 3))
	err := txEnv.Sign(privAccFrom)
	require.NoError(t, err)

	bs, err := codec.EncodeTx(txEnv)
	require.NoError(t, err)
	txEnvOut, err := codec.DecodeTx(bs)
	require.NoError(t, err)
	assert.Equal(t, txEnv, txEnvOut)
}
