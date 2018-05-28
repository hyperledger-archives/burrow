package txs

import (
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/stretchr/testify/assert"
)

func TestEncodeTxDecodeTx(t *testing.T) {
	gwc := NewGoWireCodec()
	inputAddress := acm.Address{1, 2, 3, 4, 5}
	outputAddress := acm.Address{5, 4, 3, 2, 1}
	amount := uint64(2)
	sequence := uint64(3)
	tx := &SendTx{
		Inputs: []TxInput{{
			Address:  inputAddress,
			Amount:   amount,
			Sequence: sequence,
		}},
		Outputs: []TxOutput{{
			Address: outputAddress,
			Amount:  amount,
		}},
	}
	txBytes, err := gwc.EncodeTx(tx)
	if err != nil {
		t.Fatal(err)
	}
	txOut, err := gwc.DecodeTx(txBytes)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, tx, txOut)
}

func TestEncodeTxDecodeTx_CallTx(t *testing.T) {
	gwc := NewGoWireCodec()
	inputAddress := acm.Address{1, 2, 3, 4, 5}
	amount := uint64(2)
	sequence := uint64(3)
	tx := &CallTx{
		Input: TxInput{
			Address:  inputAddress,
			Amount:   amount,
			Sequence: sequence,
		},
		GasLimit: 233,
		Fee:      2,
		Address:  nil,
		Data:     []byte("code"),
	}
	txBytes, err := gwc.EncodeTx(tx)
	if err != nil {
		t.Fatal(err)
	}
	txOut, err := gwc.DecodeTx(txBytes)
	assert.NoError(t, err, "DecodeTx error")
	assert.Equal(t, tx, txOut)
}
