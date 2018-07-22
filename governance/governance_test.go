package governance

import (
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/require"
)

func TestSerialise(t *testing.T) {
	priv := acm.GeneratePrivateAccountFromSecret("hhelo")
	tx := AlterPowerTx(priv.Address(), priv, 3242323)
	txEnv := txs.Enclose("OOh", tx)

	txe := exec.NewTxExecution(txEnv)
	bs, err := txe.Marshal()
	require.NoError(t, err)
	txeOut := new(exec.TxExecution)
	err = txeOut.Unmarshal(bs)
	require.NoError(t, err)
}
