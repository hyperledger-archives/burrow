// +build integration

package rpctransact

import (
	"context"
	"testing"

	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendTxSync(t *testing.T) {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	for i := 0; i < 2; i++ {
		txe, err := cli.SendTxSync(context.Background(), &payload.SendTx{
			Inputs: []*payload.TxInput{{
				Address: inputAddress,
				Amount:  2003,
			}},
			Outputs: []*payload.TxOutput{{
				Address: rpctest.PrivateAccounts[3].GetAddress(),
				Amount:  2003,
			}},
		})
		require.NoError(t, err)
		assert.False(t, txe.Receipt.CreatesContract)
	}
}
