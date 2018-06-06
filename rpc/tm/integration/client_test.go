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

package integration

import (
	"bytes"
	"testing"
	"time"

	"fmt"
	"strings"

	"github.com/hyperledger/burrow/binary"
	exe_events "github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/rpc"
	tm_client "github.com/hyperledger/burrow/rpc/tm/client"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/consensus/types"
	"golang.org/x/crypto/ripemd160"
)

func testWithAllClients(t *testing.T, testFunction func(*testing.T, string, tm_client.RPCClient)) {
	for clientName, client := range clients {
		testFunction(t, clientName, client)
	}
}

//--------------------------------------------------------------------------------
func TestStatus(t *testing.T) {
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		resp, err := tm_client.Status(client)
		assert.NoError(t, err)
		assert.Equal(t, genesisDoc.ChainID(), resp.NodeInfo.Network, "ChainID should match NodeInfo.Network")
	})
}

func TestBroadcastTx(t *testing.T) {
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		// Avoid duplicate Tx in mempool
		amt := hashString(clientName) % 1000
		toAddr := privateAccounts[1].Address()
		tx := makeDefaultSendTxSigned(t, client, toAddr, amt)
		receipt, err := broadcastTxAndWait(t, client, tx)
		require.NoError(t, err)
		assert.False(t, receipt.CreatesContract, "This tx should not create a contract")
		assert.NotEmpty(t, receipt.TxHash, "Failed to compute tx hash")

		buf, n, errp := new(bytes.Buffer), new(int), new(error)
		hasher := ripemd160.New()
		tx.WriteSignBytes(genesisDoc.ChainID(), buf, n, errp)
		assert.NoError(t, *errp)
		txSignBytes := buf.Bytes()
		hasher.Write(txSignBytes)
		txHashExpected := hasher.Sum(nil)

		if bytes.Compare(receipt.TxHash, txHashExpected) != 0 {
			t.Fatalf("The receipt hash '%x' does not equal the ripemd160 hash of the "+
				"transaction signed bytes calculated in the test: '%x'",
				receipt.TxHash, txHashExpected)
		}
	})
}

func TestGetAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		acc := getAccount(t, client, privateAccounts[0].Address())
		if acc == nil {
			t.Fatal("Account was nil")
		}
		if acc.Address() != privateAccounts[0].Address() {
			t.Fatalf("Failed to get correct account. Got %s, expected %s", acc.Address(),
				privateAccounts[0].Address())
		}
	})
}

func TestGetStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		amt, gasLim, fee := uint64(1100), uint64(1000), uint64(1000)
		code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
		// Call with nil address will create a contract
		tx := makeDefaultCallTx(t, client, nil, code, amt, gasLim, fee)
		receipt, err := broadcastTxAndWait(t, client, tx)
		assert.NoError(t, err)
		assert.Equal(t, true, receipt.CreatesContract, "This transaction should"+
			" create a contract")
		assert.NotEqual(t, 0, len(receipt.TxHash), "Receipt should contain a"+
			" transaction hash")
		contractAddr := receipt.ContractAddress
		assert.NotEqual(t, 0, len(contractAddr), "Transactions claims to have"+
			" created a contract but the contract address is empty")

		v := getStorage(t, client, contractAddr, []byte{0x1})
		got := binary.LeftPadWord256(v)
		expected := binary.LeftPadWord256([]byte{0x5})
		if got.Compare(expected) != 0 {
			t.Fatalf("Wrong storage value. Got %x, expected %x", got.Bytes(),
				expected.Bytes())
		}
	})
}

func TestCallCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		// add two integers and return the result
		code := []byte{0x60, 0x5, 0x60, 0x6, 0x1, 0x60, 0x0, 0x52, 0x60, 0x20, 0x60,
			0x0, 0xf3}
		data := []byte{}
		expected := []byte{0xb}
		callCode(t, client, privateAccounts[0].Address(), code, data, expected)

		// pass two ints as calldata, add, and return the result
		code = []byte{0x60, 0x0, 0x35, 0x60, 0x20, 0x35, 0x1, 0x60, 0x0, 0x52, 0x60,
			0x20, 0x60, 0x0, 0xf3}
		data = append(binary.LeftPadWord256([]byte{0x5}).Bytes(),
			binary.LeftPadWord256([]byte{0x6}).Bytes()...)
		expected = []byte{0xb}
		callCode(t, client, privateAccounts[0].Address(), code, data, expected)
	})
}

func TestCallContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		// create the contract
		amt, gasLim, fee := uint64(6969), uint64(1000), uint64(1000)
		code, _, _ := simpleContract()
		tx := makeDefaultCallTx(t, client, nil, code, amt, gasLim, fee)
		receipt, err := broadcastTxAndWait(t, client, tx)
		assert.NoError(t, err)
		if err != nil {
			t.Fatalf("Problem broadcasting transaction: %v", err)
		}
		assert.Equal(t, true, receipt.CreatesContract, "This transaction should"+
			" create a contract")
		assert.NotEqual(t, 0, len(receipt.TxHash), "Receipt should contain a"+
			" transaction hash")
		contractAddr := receipt.ContractAddress
		assert.NotEqual(t, 0, len(contractAddr), "Transactions claims to have"+
			" created a contract but the contract address is empty")

		// run a call through the contract
		data := []byte{}
		expected := []byte{0xb}
		callContract(t, client, privateAccounts[0].Address(), contractAddr, data, expected)
	})
}

func TestNameReg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	defer stopWSClient(wsc)
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		txs.MinNameRegistrationPeriod = 1

		// register a new name, check if its there
		// since entries ought to be unique and these run against different clients, we append the client
		name := "ye_old_domain_name_" + clientName
		const data = "if not now, when"
		fee := uint64(1000)
		numDesiredBlocks := uint64(2)
		amt := fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*txs.NameBaseCost(name, data)

		tx := makeDefaultNameTx(t, client, name, data, amt, fee)
		// verify the name by both using the event and by checking get_name
		subscribeAndWaitForNext(t, wsc, exe_events.EventStringNameReg(name),
			func() {
				broadcastTx(t, client, tx)
			},
			func(eventID string, resultEvent *rpc.ResultEvent) (bool, error) {

				eventDataTx := resultEvent.EventDataTx
				assert.NotNil(t, eventDataTx, "could not convert %s to EventDataTx", resultEvent)
				tx, ok := eventDataTx.Tx.(*txs.NameTx)
				if !ok {
					t.Fatalf("Could not convert %v to *NameTx", eventDataTx)
				}
				assert.Equal(t, name, tx.Name)
				assert.Equal(t, data, tx.Data)
				return true, nil
			})

		entry := getNameRegEntry(t, client, name)
		assert.Equal(t, data, entry.Data)
		assert.Equal(t, privateAccounts[0].Address(), entry.Owner)

		// update the data as the owner, make sure still there
		numDesiredBlocks = uint64(5)
		const updatedData = "these are amongst the things I wish to bestow upon " +
			"the youth of generations come: a safe supply of honey, and a better " +
			"money. For what else shall they need"
		amt = fee + numDesiredBlocks*txs.NameByteCostMultiplier*
			txs.NameBlockCostMultiplier*txs.NameBaseCost(name, updatedData)
		tx = makeDefaultNameTx(t, client, name, updatedData, amt, fee)
		broadcastTxAndWait(t, client, tx)
		entry = getNameRegEntry(t, client, name)

		assert.Equal(t, updatedData, entry.Data)

		// try to update as non owner, should fail
		tx = txs.NewNameTxWithSequence(privateAccounts[1].PublicKey(), name, "never mind", amt, fee,
			getSequence(t, client, privateAccounts[1].Address())+1)
		tx.Sign(genesisDoc.ChainID(), privateAccounts[1])

		_, err := tm_client.BroadcastTx(client, tx)
		assert.Error(t, err, "Expected error when updating someone else's unexpired"+
			" name registry entry")
		if err != nil {
			assert.Contains(t, err.Error(), "permission denied", "Error should be "+
				"permission denied")
		}

		// Wait a couple of blocks to make sure name registration expires
		waitNBlocks(t, wsc, 5)

		//now the entry should be expired, so we can update as non owner
		const data2 = "this is not my beautiful house"
		tx = txs.NewNameTxWithSequence(privateAccounts[1].PublicKey(), name, data2, amt, fee,
			getSequence(t, client, privateAccounts[1].Address())+1)
		tx.Sign(genesisDoc.ChainID(), privateAccounts[1])

		//_, err = tm_client.BroadcastTx(client, tx)
		require.NoError(t, subscribeAndWaitForNext(t, wsc, exe_events.EventStringNameReg(name),
			func() {
				_, err = tm_client.BroadcastTx(client, tx)
				assert.NoError(t, err, "Should be able to update a previously expired name"+
					" registry entry as a different address")
			},
			func(eventID string, resultEvent *rpc.ResultEvent) (bool, error) {
				return true, nil
			}))
		entry = getNameRegEntry(t, client, name)
		assert.Equal(t, data2, entry.Data)
		assert.Equal(t, privateAccounts[1].Address(), entry.Owner)
	})
}

func TestWaitBlocks(t *testing.T) {
	wsc := newWSClient()
	defer stopWSClient(wsc)
	waitNBlocks(t, wsc, 5)
}

func TestBlockchainInfo(t *testing.T) {
	wsc := newWSClient()
	defer stopWSClient(wsc)
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		// wait a mimimal number of blocks to ensure that the later query for block
		// headers has a non-trivial length
		nBlocks := 4
		waitNBlocks(t, wsc, nBlocks)

		resp, err := tm_client.ListBlocks(client, 0, 0)
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
		resp, err = tm_client.ListBlocks(client, 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resp.BlockMetas),
			"Should see 2 BlockMetas after extracting 2 blocks")
	})
}

func TestListUnconfirmedTxs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	defer stopWSClient(wsc)
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		amt, gasLim, fee := uint64(1100), uint64(1000), uint64(1000)
		code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
		// Call with nil address will create a contract
		tx := txs.Wrap(makeDefaultCallTx(t, client, nil, code, amt, gasLim, fee))
		txChan := make(chan []txs.Body)

		// We want to catch the Tx in mempool before it gets reaped by tendermint
		// consensus. We should be able to do this almost always if we broadcast our
		// transaction immediately after a block has been committed. There is about
		// 1 second between blocks, and we will have the lock on Reap
		// So we wait for a block here
		waitNBlocks(t, wsc, 1)

		go func() {
			for {
				resp, err := tm_client.ListUnconfirmedTxs(client, -1)
				if resp != nil {

				}
				require.NoError(t, err)
				if resp.NumTxs > 0 {
					txChan <- resp.Txs
				}
			}
		}()

		require.NoError(t, runThenWaitForBlock(t, wsc, nextBlockPredicateFn(), func() {
			broadcastTx(t, client, tx)
			select {
			case <-time.After(time.Second * timeoutSeconds * 10):
				t.Fatal("Timeout out waiting for unconfirmed transactions to appear")
			case transactions := <-txChan:
				assert.Len(t, transactions, 1, "There should only be a single transaction in the "+
					"mempool during this test (previous txs should have made it into a block)")
				assert.Contains(t, transactions, tx, "Transaction should be returned by ListUnconfirmedTxs")
			}
		}))
	})
}

func TestGetBlock(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		waitNBlocks(t, wsc, 3)
		resp, err := tm_client.GetBlock(client, 2)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), resp.Block.Height)
		assert.Equal(t, int64(2), resp.BlockMeta.Header.Height)
	})
}

func TestListValidators(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		waitNBlocks(t, wsc, 3)
		resp, err := tm_client.ListValidators(client)
		assert.NoError(t, err)
		assert.Len(t, resp.BondedValidators, 1)
		validator := resp.BondedValidators[0]
		assert.Equal(t, genesisDoc.Validators[0].PublicKey, validator.PublicKey)
	})
}

func TestDumpConsensusState(t *testing.T) {
	wsc := newWSClient()
	startTime := time.Now()
	testWithAllClients(t, func(t *testing.T, clientName string, client tm_client.RPCClient) {
		waitNBlocks(t, wsc, 3)
		resp, err := tm_client.DumpConsensusState(client)
		assert.NoError(t, err)
		assert.NotZero(t, startTime)
		assert.Equal(t, fmt.Sprintf("/0/%d", types.RoundStepNewHeight),
			strings.TrimLeft(resp.RoundState.HeightRoundStep, "0123456789"))
	})
}
