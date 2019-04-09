// +build integration

package rpctransact

import (
	"context"
	"testing"

	"github.com/hyperledger/burrow/integration"

	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendTx(t *testing.T) {
	t.Parallel()
	kern, shutdown := integration.RunNode(t, rpctest.GenesisDoc, rpctest.PrivateAccounts)
	defer shutdown()

	t.Run("Sync", func(t *testing.T) {
		cli := rpctest.NewTransactClient(t, kern.GRPCListenAddress().String())
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
	})

	t.Run("Async", func(t *testing.T) {
		cli := rpctest.NewTransactClient(t, kern.GRPCListenAddress().String())
		numSends := 1000
		expecter := rpctest.ExpectTxs(kern.Emitter, "SendTxAsync")
		for i := 0; i < numSends; i++ {
			receipt, err := cli.SendTxAsync(context.Background(), &payload.SendTx{
				Inputs: []*payload.TxInput{{
					Address: inputAddress,
					Amount:  2003,
				}},
				Outputs: []*payload.TxOutput{{
					Address: rpctest.PrivateAccounts[3].GetAddress(),
					Amount:  2003,
				}},
			})
			expecter.Expect(receipt.TxHash)
			require.NoError(t, err)
			assert.False(t, receipt.CreatesContract)
		}
		expecter.AssertCommitted(t)
	})
}
