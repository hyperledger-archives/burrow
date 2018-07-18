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

	"github.com/hyperledger/burrow/execution/exec"
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
var inputAddress = inputAccount.Address()

func TestBroadcastTxLocallySigned(t *testing.T) {
	qcli := rpctest.NewQueryClient(t, testConfig.RPC.GRPC.ListenAddress)
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
			Address: rpctest.PrivateAccounts[1].Address(),
			Amount:  amount,
		}},
	})
	require.NoError(t, txEnv.Sign(inputAccount))

	// Test subscribing to transaction before sending it
	ch := make(chan *exec.TxExecution)
	go func() {
		ecli := rpctest.NewExecutionEventsClient(t, testConfig.RPC.GRPC.ListenAddress)
		txe, err := ecli.GetTx(context.Background(), &rpcevents.GetTxRequest{
			TxHash: txEnv.Tx.Hash(),
			Wait:   true,
		})
		require.NoError(t, err)
		ch <- txe
	}()

	// Make it wait
	time.Sleep(time.Second)

	// No broadcast
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	receipt, err := cli.BroadcastTxAsync(context.Background(), &rpctransact.TxEnvelopeParam{Envelope: txEnv})
	require.NoError(t, err)
	assert.False(t, receipt.CreatesContract, "This tx should not create a contract")
	require.NotEmpty(t, receipt.TxHash, "Failed to compute tx hash")
	assert.Equal(t, txEnv.Tx.Hash(), receipt.TxHash)

	txe := <-ch
	require.True(t, len(txe.Events) > 0)
	assert.NotNil(t, txe.Events[0].Input)
}

func TestFormulateTx(t *testing.T) {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	txEnv, err := cli.FormulateTx(context.Background(), &rpctransact.PayloadParam{
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
		"\"Payload\":{\"Input\":{\"Address\":\"4A6DFB649EF0D50780998A686BD69AB175C08E26\",\"Amount\":230},"+
		"\"Data\":\"0203060403\"}}", rpctest.GenesisDoc.ChainID())) {
		fmt.Println(string(bs))
	}
}
