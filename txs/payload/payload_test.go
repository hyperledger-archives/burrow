package payload

import (
	"encoding/json"
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomTypes(t *testing.T) {
	address := crypto.Address{1, 2, 3}
	callTx := &CallTx{
		Input: &TxInput{
			Address:  crypto.Address{1, 2},
			Amount:   2,
			Sequence: 0,
		},
		Address: &address,
	}
	bs, err := encoding.Encode(callTx)
	require.NoError(t, err)
	callTxOut := new(CallTx)
	err = encoding.Decode(bs, callTxOut)
	require.NoError(t, err)
	assert.Equal(t, jsonString(t, callTx), jsonString(t, callTxOut))
}

func jsonString(t testing.TB, conf interface{}) string {
	bs, err := json.MarshalIndent(conf, "", "  ")
	require.NoError(t, err, "must be able to convert interface to string for comparison")
	return string(bs)
}
