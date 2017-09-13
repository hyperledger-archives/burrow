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

package test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"golang.org/x/crypto/ripemd160"

	consensus_types "github.com/hyperledger/burrow/consensus/types"
	burrow_client "github.com/hyperledger/burrow/rpc/tendermint/client"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/word256"

	"github.com/stretchr/testify/assert"
)

// When run with `-test.short` we only run:
// TestHTTPStatus, TestHTTPBroadcast, TestJSONStatus, TestJSONBroadcast, TestWSConnect, TestWSSend

// Note: the reason that we have tests implemented in tests.go is I believe
// due to weirdness with go-wire's interface registration, and those global
// registrations not being available within a *_test.go runtime context.
func testWithAllClients(t *testing.T,
	testFunction func(*testing.T, string, burrow_client.RPCClient)) {
	for clientName, client := range clients {
		testFunction(t, clientName, client)
	}
}

//--------------------------------------------------------------------------------
func TestStatus(t *testing.T) {
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		resp, err := burrow_client.Status(client)
		assert.NoError(t, err)
		fmt.Println(resp)
		if resp.NodeInfo.Network != chainID {
			t.Fatal(fmt.Errorf("ChainID mismatch: got %s expected %s",
				resp.NodeInfo.Network, chainID))
		}
	})
}

func TestBroadcastTx(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		// Avoid duplicate Tx in mempool
		amt := hashString(clientName) % 1000
		toAddr := users[1].Address
		tx := makeDefaultSendTxSigned(t, client, toAddr, amt)
		receipt, err := broadcastTxAndWaitForBlock(t, client, wsc, tx)
		assert.NoError(t, err)
		if receipt.CreatesContract > 0 {
			t.Fatal("This tx does not create a contract")
		}
		if len(receipt.TxHash) == 0 {
			t.Fatal("Failed to compute tx hash")
		}
		n, errp := new(int), new(error)
		buf := new(bytes.Buffer)
		hasher := ripemd160.New()
		tx.WriteSignBytes(chainID, buf, n, errp)
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
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		acc := getAccount(t, client, users[0].Address)
		if acc == nil {
			t.Fatal("Account was nil")
		}
		if bytes.Compare(acc.Address, users[0].Address) != 0 {
			t.Fatalf("Failed to get correct account. Got %x, expected %x", acc.Address,
				users[0].Address)
		}
	})
}

func TestGetStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	defer func() {
		wsc.Stop()
	}()
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		eid := txs.EventStringNewBlock()
		subscribe(t, wsc, eid)
		defer func() {
			unsubscribe(t, wsc, eid)
		}()

		amt, gasLim, fee := int64(1100), int64(1000), int64(1000)
		code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
		// Call with nil address will create a contract
		tx := makeDefaultCallTx(t, client, nil, code, amt, gasLim, fee)
		receipt, err := broadcastTxAndWaitForBlock(t, client, wsc, tx)
		assert.NoError(t, err)
		assert.Equal(t, uint8(1), receipt.CreatesContract, "This transaction should"+
			" create a contract")
		assert.NotEqual(t, 0, len(receipt.TxHash), "Receipt should contain a"+
			" transaction hash")
		contractAddr := receipt.ContractAddr
		assert.NotEqual(t, 0, len(contractAddr), "Transactions claims to have"+
			" created a contract but the contract address is empty")

		v := getStorage(t, client, contractAddr, []byte{0x1})
		got := word256.LeftPadWord256(v)
		expected := word256.LeftPadWord256([]byte{0x5})
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

	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		// add two integers and return the result
		code := []byte{0x60, 0x5, 0x60, 0x6, 0x1, 0x60, 0x0, 0x52, 0x60, 0x20, 0x60,
			0x0, 0xf3}
		data := []byte{}
		expected := []byte{0xb}
		callCode(t, client, users[0].PubKey.Address(), code, data, expected)

		// pass two ints as calldata, add, and return the result
		code = []byte{0x60, 0x0, 0x35, 0x60, 0x20, 0x35, 0x1, 0x60, 0x0, 0x52, 0x60,
			0x20, 0x60, 0x0, 0xf3}
		data = append(word256.LeftPadWord256([]byte{0x5}).Bytes(),
			word256.LeftPadWord256([]byte{0x6}).Bytes()...)
		expected = []byte{0xb}
		callCode(t, client, users[0].PubKey.Address(), code, data, expected)
	})
}

func TestCallContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	defer func() {
		wsc.Stop()
	}()
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		eid := txs.EventStringNewBlock()
		subscribe(t, wsc, eid)
		defer func() {
			unsubscribe(t, wsc, eid)
		}()

		// create the contract
		amt, gasLim, fee := int64(6969), int64(1000), int64(1000)
		code, _, _ := simpleContract()
		tx := makeDefaultCallTx(t, client, nil, code, amt, gasLim, fee)
		receipt, err := broadcastTxAndWaitForBlock(t, client, wsc, tx)
		assert.NoError(t, err)
		if err != nil {
			t.Fatalf("Problem broadcasting transaction: %v", err)
		}
		assert.Equal(t, uint8(1), receipt.CreatesContract, "This transaction should"+
			" create a contract")
		assert.NotEqual(t, 0, len(receipt.TxHash), "Receipt should contain a"+
			" transaction hash")
		contractAddr := receipt.ContractAddr
		assert.NotEqual(t, 0, len(contractAddr), "Transactions claims to have"+
			" created a contract but the contract address is empty")

		// run a call through the contract
		data := []byte{}
		expected := []byte{0xb}
		callContract(t, client, users[0].PubKey.Address(), contractAddr, data, expected)
	})
}

func TestNameReg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {

		txs.MinNameRegistrationPeriod = 1

		// register a new name, check if its there
		// since entries ought to be unique and these run against different clients, we append the client
		name := "ye_old_domain_name_" + clientName
		const data = "if not now, when"
		fee := int64(1000)
		numDesiredBlocks := int64(2)
		amt := fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*txs.NameBaseCost(name, data)

		tx := makeDefaultNameTx(t, client, name, data, amt, fee)
		// verify the name by both using the event and by checking get_name
		subscribeAndWaitForNext(t, wsc, txs.EventStringNameReg(name),
			func() {
				broadcastTxAndWaitForBlock(t, client, wsc, tx)
			},
			func(eid string, eventData txs.EventData) (bool, error) {
				eventDataTx := asEventDataTx(t, eventData)
				tx, ok := eventDataTx.Tx.(*txs.NameTx)
				if !ok {
					t.Fatalf("Could not convert %v to *NameTx", eventDataTx)
				}
				assert.Equal(t, name, tx.Name)
				assert.Equal(t, data, tx.Data)
				return true, nil
			})
		mempoolCount = 0

		entry := getNameRegEntry(t, client, name)
		assert.Equal(t, data, entry.Data)
		assert.Equal(t, users[0].Address, entry.Owner)

		// update the data as the owner, make sure still there
		numDesiredBlocks = int64(5)
		const updatedData = "these are amongst the things I wish to bestow upon " +
			"the youth of generations come: a safe supply of honey, and a better " +
			"money. For what else shall they need"
		amt = fee + numDesiredBlocks*txs.NameByteCostMultiplier*
			txs.NameBlockCostMultiplier*txs.NameBaseCost(name, updatedData)
		tx = makeDefaultNameTx(t, client, name, updatedData, amt, fee)
		broadcastTxAndWaitForBlock(t, client, wsc, tx)
		mempoolCount = 0
		entry = getNameRegEntry(t, client, name)

		assert.Equal(t, updatedData, entry.Data)

		// try to update as non owner, should fail
		tx = txs.NewNameTxWithNonce(users[1].PubKey, name, "never mind", amt, fee,
			getNonce(t, client, users[1].Address)+1)
		tx.Sign(chainID, users[1])

		_, err := broadcastTxAndWaitForBlock(t, client, wsc, tx)
		assert.Error(t, err, "Expected error when updating someone else's unexpired"+
			" name registry entry")
		if err != nil {
			assert.Contains(t, err.Error(), "permission denied", "Error should be "+
				"permission denied")
		}

		// Wait a couple of blocks to make sure name registration expires
		waitNBlocks(t, wsc, 3)

		//now the entry should be expired, so we can update as non owner
		const data2 = "this is not my beautiful house"
		tx = txs.NewNameTxWithNonce(users[1].PubKey, name, data2, amt, fee,
			getNonce(t, client, users[1].Address)+1)
		tx.Sign(chainID, users[1])
		_, err = broadcastTxAndWaitForBlock(t, client, wsc, tx)
		assert.NoError(t, err, "Should be able to update a previously expired name"+
			" registry entry as a different address")
		mempoolCount = 0
		entry = getNameRegEntry(t, client, name)
		assert.Equal(t, data2, entry.Data)
		assert.Equal(t, users[1].Address, entry.Owner)
	})
}

func TestBlockchainInfo(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		// wait a mimimal number of blocks to ensure that the later query for block
		// headers has a non-trivial length
		nBlocks := 4
		waitNBlocks(t, wsc, nBlocks)

		resp, err := burrow_client.BlockchainInfo(client, 0, 0)
		if err != nil {
			t.Fatalf("Failed to get blockchain info: %v", err)
		}
		lastBlockHeight := resp.LastHeight
		nMetaBlocks := len(resp.BlockMetas)
		assert.True(t, nMetaBlocks <= lastBlockHeight,
			"Logically number of block metas should be equal or less than block height.")
		assert.True(t, nBlocks <= len(resp.BlockMetas),
			"Should see at least 4 BlockMetas after waiting for 4 blocks")
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
		resp, err = burrow_client.BlockchainInfo(client, 1, 2)
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
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		amt, gasLim, fee := int64(1100), int64(1000), int64(1000)
		code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
		// Call with nil address will create a contract
		tx := makeDefaultCallTx(t, client, []byte{}, code, amt, gasLim, fee)
		txChan := make(chan []txs.Tx)

		// We want to catch the Tx in mempool before it gets reaped by tendermint
		// consensus. We should be able to do this almost always if we broadcast our
		// transaction immediately after a block has been committed. There is about
		// 1 second between blocks, and we will have the lock on Reap
		// So we wait for a block here
		waitNBlocks(t, wsc, 1)

		go func() {
			for {
				resp, err := burrow_client.ListUnconfirmedTxs(client)
				assert.NoError(t, err)
				if resp.N > 0 {
					txChan <- resp.Txs
				}
			}
		}()

		runThenWaitForBlock(t, wsc, nextBlockPredicateFn(), func() {
			broadcastTx(t, client, tx)
			select {
			case <-time.After(time.Second * timeoutSeconds):
				t.Fatal("Timeout out waiting for unconfirmed transactions to appear")
			case transactions := <-txChan:
				assert.Len(t, transactions, 1,
					"There should only be a single transaction in the mempool during "+
						"this test (previous txs should have made it into a block)")
				assert.Contains(t, transactions, tx,
					"Transaction should be returned by ListUnconfirmedTxs")
			}
		})
	})
}

func TestGetBlock(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		waitNBlocks(t, wsc, 3)
		resp, err := burrow_client.GetBlock(client, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, resp.Block.Height)
		assert.Equal(t, 2, resp.BlockMeta.Header.Height)
	})
}

func TestListValidators(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		waitNBlocks(t, wsc, 3)
		resp, err := burrow_client.ListValidators(client)
		assert.NoError(t, err)
		assert.Len(t, resp.BondedValidators, 1)
		validator := resp.BondedValidators[0].(*consensus_types.TendermintValidator)
		assert.Equal(t, genesisDoc.Validators[0].PubKey, validator.PubKey)
	})
}

func TestDumpConsensusState(t *testing.T) {
	wsc := newWSClient()
	testWithAllClients(t, func(t *testing.T, clientName string, client burrow_client.RPCClient) {
		waitNBlocks(t, wsc, 3)
		resp, err := burrow_client.DumpConsensusState(client)
		assert.NoError(t, err)
		startTime := resp.ConsensusState.StartTime
		// TODO: uncomment lines involving commitTime when
		// https://github.com/tendermint/tendermint/issues/277 is fixed in Tendermint
		//commitTime := resp.ConsensusState.CommitTime
		assert.NotZero(t, startTime)
		//assert.NotZero(t, commitTime)
		//assert.True(t, commitTime.Unix() > startTime.Unix(),
		//	"Commit time %v should be later than start time %v", commitTime, startTime)
		assert.Equal(t, uint8(1), resp.ConsensusState.Step)
	})
}

func asEventDataTx(t *testing.T, eventData txs.EventData) txs.EventDataTx {
	eventDataTx, ok := eventData.(txs.EventDataTx)
	if !ok {
		t.Fatalf("Expected eventData to be EventDataTx was %v", eventData)
	}
	return eventDataTx
}
