// +build integration

// Space above here matters
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

package rpctransact

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/burrow/integration"

	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var inputAccount = rpctest.PrivateAccounts[0]
var inputAddress = inputAccount.GetAddress()

func TestTransactServer(t *testing.T) {
	t.Parallel()
	kern, shutdown := integration.RunNode(t, rpctest.GenesisDoc, rpctest.PrivateAccounts)
	defer shutdown()

	t.Run("InputAccountPublicKeySet", func(t *testing.T) {
		input := rpctest.PrivateAccounts[9]
		tcli := rpctest.NewTransactClient(t, kern.GRPCListenAddress().String())
		qcli := rpctest.NewQueryClient(t, kern.GRPCListenAddress().String())
		acc, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: input.GetAddress()})
		require.NoError(t, err)

		// Account PublicKey should be initially unset
		assert.False(t, acc.PublicKey.IsSet())

		// Sign with this account - should set public key
		_, err = rpctest.CreateContract(tcli, input.GetAddress(), solidity.Bytecode_StrangeLoop, nil)
		require.NoError(t, err)
		acc, err = qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: input.GetAddress()})

		// Check public key set
		require.NoError(t, err)
		assert.True(t, acc.PublicKey.IsSet())
		assert.Equal(t, input.GetPublicKey(), acc.PublicKey)
	})

	t.Run("BroadcastTxLocallySigned", func(t *testing.T) {
		qcli := rpctest.NewQueryClient(t, kern.GRPCListenAddress().String())
		acc, err := qcli.GetAccount(context.Background(), &rpcquery.GetAccountParam{
			Address: inputAddress,
		})
		require.NoError(t, err)
		amount := uint64(2123)
		txEnv := txs.Enclose(rpctest.GenesisDoc.ChainID(), &payload.SendTx{
			Inputs: []*payload.TxInput{{
				Address:  inputAddress,
				Sequence: acc.Sequence + 1,
				Amount:   amount,
			}},
			Outputs: []*payload.TxOutput{{
				Address: rpctest.PrivateAccounts[1].GetAddress(),
				Amount:  amount,
			}},
		})
		require.NoError(t, txEnv.Sign(inputAccount))

		// Test subscribing to transaction before sending it
		ch := make(chan *exec.TxExecution)
		go func() {
			ecli := rpctest.NewExecutionEventsClient(t, kern.GRPCListenAddress().String())
			txe, err := ecli.Tx(context.Background(), &rpcevents.TxRequest{
				TxHash: txEnv.Tx.Hash(),
				Wait:   true,
			})
			require.NoError(t, err)
			ch <- txe
		}()

		// Make it wait
		time.Sleep(time.Second)

		// No broadcast
		cli := rpctest.NewTransactClient(t, kern.GRPCListenAddress().String())
		receipt, err := cli.BroadcastTxAsync(context.Background(), &rpctransact.TxEnvelopeParam{Envelope: txEnv})
		require.NoError(t, err)
		assert.False(t, receipt.CreatesContract, "This tx should not create a contract")
		require.NotEmpty(t, receipt.TxHash, "Failed to compute tx hash")
		assert.Equal(t, txEnv.Tx.Hash(), receipt.TxHash)

		txe := <-ch
		require.True(t, len(txe.Events) > 0)
		assert.NotNil(t, txe.Events[0].Input)
	})

	t.Run("FormulateTx", func(t *testing.T) {
		cli := rpctest.NewTransactClient(t, kern.GRPCListenAddress().String())
		txEnv, err := cli.FormulateTx(context.Background(), &payload.Any{
			CallTx: &payload.CallTx{
				Input: &payload.TxInput{
					Address: inputAddress,
					Amount:  230,
				},
				Data: []byte{2, 3, 6, 4, 3},
			},
		})
		require.NoError(t, err)
		bs, err := txEnv.Marshal()
		require.NoError(t, err)
		// We should see the sign bytes embedded
		if !assert.Contains(t, string(bs), fmt.Sprintf("{\"ChainID\":\"%s\",\"Type\":\"CallTx\","+
			"\"Payload\":{\"Input\":{\"Address\":\"E80BB91C2F0F4C3C39FC53E89BF8416B219BE6E4\",\"Amount\":230},"+
			"\"Data\":\"0203060403\"}}", rpctest.GenesisDoc.ChainID())) {
			fmt.Println(string(bs))
		}
	})
}
