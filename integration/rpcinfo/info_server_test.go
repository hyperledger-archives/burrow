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

package rpcinfo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/rpcinfo/infoclient"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ctypes "github.com/tendermint/tendermint/consensus/types"
)

const timeout = 5 * time.Second

func testWithAllClients(t *testing.T, testFunction func(*testing.T, string, infoclient.RPCClient)) {
	for clientName, client := range clients {
		testFunction(t, clientName, client)
	}
}

//--------------------------------------------------------------------------------
func TestStatus(t *testing.T) {
	testWithAllClients(t, func(t *testing.T, clientName string, client infoclient.RPCClient) {
		resp, err := infoclient.Status(client)
		require.NoError(t, err)
		assert.Equal(t, "node_001", resp.NodeInfo.Moniker)
		assert.Equal(t, rpctest.GenesisDoc.ChainID(), resp.NodeInfo.Network,
			"ChainID should match NodeInfo.Network")
	})
}

func TestAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testWithAllClients(t, func(t *testing.T, clientName string, client infoclient.RPCClient) {
		acc := rpctest.GetAccount(t, client, rpctest.PrivateAccounts[0].Address())
		if acc == nil {
			t.Fatal("Account was nil")
		}
		if acc.Address() != rpctest.PrivateAccounts[0].Address() {
			t.Fatalf("Failed to get correct account. Got %s, expected %s", acc.Address(),
				rpctest.PrivateAccounts[0].Address())
		}
	})
}

func TestStorage(t *testing.T) {
	testWithAllClients(t, func(t *testing.T, clientName string, client infoclient.RPCClient) {
		amt, gasLim, fee := uint64(1100), uint64(1000), uint64(1000)
		code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
		// Call with nil address will create a contract
		tx := rpctest.MakeDefaultCallTx(t, client, nil, code, amt, gasLim, fee)
		txe := broadcastTxSync(t, tx)
		assert.Equal(t, true, txe.Receipt.CreatesContract, "This transaction should"+
			" create a contract")
		assert.NotEqual(t, 0, len(txe.TxHash), "Receipt should contain a"+
			" transaction hash")
		contractAddr := txe.Receipt.ContractAddress
		assert.NotEqual(t, 0, len(contractAddr), "Transactions claims to have"+
			" created a contract but the contract address is empty")

		v := rpctest.GetStorage(t, client, contractAddr, []byte{0x1})
		got := binary.LeftPadWord256(v)
		expected := binary.LeftPadWord256([]byte{0x5})
		if got.Compare(expected) != 0 {
			t.Fatalf("Wrong storage value. Got %x, expected %x", got.Bytes(),
				expected.Bytes())
		}
	})
}

func TestBlock(t *testing.T) {
	waitNBlocks(t, 1)
	testWithAllClients(t, func(t *testing.T, clientName string, client infoclient.RPCClient) {
		res, err := infoclient.Block(client, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), res.Block.Height)
	})
}

func TestWaitBlocks(t *testing.T) {
	waitNBlocks(t, 5)
}

func TestBlockchainInfo(t *testing.T) {
	testWithAllClients(t, func(t *testing.T, clientName string, client infoclient.RPCClient) {
		// wait a mimimal number of blocks to ensure that the later query for block
		// headers has a non-trivial length
		nBlocks := 4
		waitNBlocks(t, nBlocks)

		resp, err := infoclient.Blocks(client, 1, 0)
		if err != nil {
			t.Fatalf("Failed to get blockchain info: %v", err)
		}
		lastBlockHeight := resp.LastHeight
		nMetaBlocks := len(resp.BlockMetas)
		assert.True(t, uint64(nMetaBlocks) <= lastBlockHeight,
			"Logically number of block metas should be equal or less than block height.")
		assert.True(t, nBlocks <= len(resp.BlockMetas),
			"Should see at least %v BlockMetas after waiting for %v blocks but saw %v",
			nBlocks, nBlocks, len(resp.BlockMetas))
		// For the maximum number (default to 20) of retrieved block headers,
		// check that they correctly chain to each other.
		lastBlockHash := resp.BlockMetas[nMetaBlocks-1].Header.Hash()
		for i := nMetaBlocks - 2; i >= 0; i-- {
			// the blockhash in header of height h should be identical to the hash
			// in the LastBlockID of the header of block height h+1.
			assert.Equal(t, lastBlockHash, resp.BlockMetas[i].Header.LastBlockID.Hash,
				"Blockchain should be a hash tree!")
			lastBlockHash = resp.BlockMetas[i].Header.Hash()
		}

		// Now retrieve only two blockheaders (h=1, and h=2) and check that we got
		// two results.
		resp, err = infoclient.Blocks(client, 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resp.BlockMetas),
			"Should see 2 BlockMetas after extracting 2 blocks")
	})
}

func TestUnconfirmedTxs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testWithAllClients(t, func(t *testing.T, clientName string, client infoclient.RPCClient) {
		amt, gasLim, fee := uint64(1100), uint64(1000), uint64(1000)
		code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
		// Call with nil address will create a contract
		txEnv := rpctest.MakeDefaultCallTx(t, client, nil, code, amt, gasLim, fee)
		txChan := make(chan []*txs.Envelope)

		// We want to catch the Tx in mempool before it gets reaped by tendermint
		// consensus. We should be able to do this almost always if we broadcast our
		// transaction immediately after a block has been committed. There is about
		// 1 second between blocks, and we will have the lock on Reap
		// So we wait for a block here
		waitNBlocks(t, 1)

		go func() {
			for {
				resp, err := infoclient.UnconfirmedTxs(client, -1)
				if err != nil {
					// We get an error on exit
					return
				}
				if resp.NumTxs > 0 {
					txChan <- resp.Txs
				}
			}
		}()

		broadcastTxSync(t, txEnv)
		select {
		case <-time.After(time.Second * timeout):
			t.Fatal("Timeout out waiting for unconfirmed transactions to appear")
		case transactions := <-txChan:
			assert.Len(t, transactions, 1, "There should only be a single transaction in the "+
				"mempool during this test (previous txs should have made it into a block)")
			assert.Contains(t, transactions, txEnv, "Transaction should be returned by ListUnconfirmedTxs")
		}
	})
}

func TestValidators(t *testing.T) {
	testWithAllClients(t, func(t *testing.T, clientName string, client infoclient.RPCClient) {
		resp, err := infoclient.Validators(client)
		assert.NoError(t, err)
		assert.Len(t, resp.BondedValidators, 1)
		validator := resp.BondedValidators[0]
		assert.Equal(t, rpctest.GenesisDoc.Validators[0].PublicKey, validator.PublicKey)
	})
}

func TestConsensus(t *testing.T) {
	testWithAllClients(t, func(t *testing.T, clientName string, client infoclient.RPCClient) {
		resp, err := infoclient.Consensus(client)
		require.NoError(t, err)

		// Now I do a special dance... because the votes section of RoundState has will Marshal but not Unmarshal yet
		// TODO: put in a PR in tendermint to fix thiss
		rawMap := make(map[string]json.RawMessage)
		err = json.Unmarshal(resp.RoundState, &rawMap)
		require.NoError(t, err)
		delete(rawMap, "votes")

		bs, err := json.Marshal(rawMap)
		require.NoError(t, err)

		cdc := rpc.NewAminoCodec()
		rs := new(ctypes.RoundState)
		err = cdc.UnmarshalJSON(bs, rs)
		require.NoError(t, err)

		assert.Equal(t, rs.Validators.Validators[0].Address, rs.Validators.Proposer.Address)
	})
}

func waitNBlocks(t testing.TB, n int) {
	subID := event.GenSubID()
	ch, err := kern.Emitter.Subscribe(context.Background(), subID, exec.QueryForBlockExecution(), 10)
	require.NoError(t, err)
	defer kern.Emitter.UnsubscribeAll(context.Background(), subID)
	for i := 0; i < n; i++ {
		<-ch
	}
}

func broadcastTxSync(t testing.TB, txEnv *txs.Envelope) *exec.TxExecution {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	txe, err := cli.BroadcastTxSync(context.Background(), &rpctransact.TxEnvelopeParam{
		Envelope: txEnv,
	})
	require.NoError(t, err)
	return txe
}
