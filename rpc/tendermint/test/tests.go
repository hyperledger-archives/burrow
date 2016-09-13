package test

import (
	"bytes"
	"fmt"
	"testing"

	edbcli "github.com/eris-ltd/eris-db/rpc/tendermint/client"
	core_types "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	"github.com/eris-ltd/eris-db/txs"
	"github.com/stretchr/testify/assert"

	"time"

	tm_common "github.com/tendermint/go-common"
	"golang.org/x/crypto/ripemd160"
)

func testStatus(t *testing.T, typ string) {
	client := clients[typ]
	resp, err := edbcli.Status(client)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
	if resp.NodeInfo.Network != chainID {
		t.Fatal(fmt.Errorf("ChainID mismatch: got %s expected %s",
			resp.NodeInfo.Network, chainID))
	}
}

func testGetAccount(t *testing.T, typ string) {
	acc := getAccount(t, typ, user[0].Address)
	if acc == nil {
		t.Fatal("Account was nil")
	}
	if bytes.Compare(acc.Address, user[0].Address) != 0 {
		t.Fatalf("Failed to get correct account. Got %x, expected %x", acc.Address, user[0].Address)
	}
}

func testOneSignTx(t *testing.T, typ string, addr []byte, amt int64) {
	tx := makeDefaultSendTx(t, typ, addr, amt)
	tx2 := signTx(t, typ, tx, user[0])
	tx2hash := txs.TxHash(chainID, tx2)
	tx.SignInput(chainID, 0, user[0])
	txhash := txs.TxHash(chainID, tx)
	if bytes.Compare(txhash, tx2hash) != 0 {
		t.Fatal("Got different signatures for signing via rpc vs tx_utils")
	}

	tx_ := signTx(t, typ, tx, user[0])
	tx = tx_.(*txs.SendTx)
	checkTx(t, user[0].Address, user[0], tx)
}

func testBroadcastTx(t *testing.T, typ string) {
	amt := int64(100)
	toAddr := user[1].Address
	tx := makeDefaultSendTxSigned(t, typ, toAddr, amt)
	receipt := broadcastTx(t, typ, tx)
	if receipt.CreatesContract > 0 {
		t.Fatal("This tx does not create a contract")
	}
	if len(receipt.TxHash) == 0 {
		t.Fatal("Failed to compute tx hash")
	}
	n, err := new(int), new(error)
	buf := new(bytes.Buffer)
	hasher := ripemd160.New()
	tx.WriteSignBytes(chainID, buf, n, err)
	// [Silas] Currently tx.TxHash uses go-wire, we we drop that we can drop the prefix here
	goWireBytes := append([]byte{0x01, 0xcf}, buf.Bytes()...)
	hasher.Write(goWireBytes)
	txHashExpected := hasher.Sum(nil)
	if bytes.Compare(receipt.TxHash, txHashExpected) != 0 {
		t.Fatalf("The receipt hash '%x' does not equal the ripemd160 hash of the "+
			"transaction signed bytes calculated in the test: '%x'",
			receipt.TxHash, txHashExpected)
	}
}

func testGetStorage(t *testing.T, typ string) {
	wsc := newWSClient(t)
	eid := txs.EventStringNewBlock()
	subscribe(t, wsc, eid)
	defer func() {
		unsubscribe(t, wsc, eid)
		wsc.Stop()
	}()

	amt, gasLim, fee := int64(1100), int64(1000), int64(1000)
	code := []byte{0x60, 0x5, 0x60, 0x1, 0x55}
	// Call with nil address will create a contract
	tx := makeDefaultCallTx(t, typ, nil, code, amt, gasLim, fee)
	receipt, err := broadcastTxAndWaitForBlock(t, typ, wsc, tx)
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

	v := getStorage(t, typ, contractAddr, []byte{0x1})
	got := tm_common.LeftPadWord256(v)
	expected := tm_common.LeftPadWord256([]byte{0x5})
	if got.Compare(expected) != 0 {
		t.Fatalf("Wrong storage value. Got %x, expected %x", got.Bytes(),
			expected.Bytes())
	}
}

func testCallCode(t *testing.T, typ string) {
	client := clients[typ]

	// add two integers and return the result
	code := []byte{0x60, 0x5, 0x60, 0x6, 0x1, 0x60, 0x0, 0x52, 0x60, 0x20, 0x60,
		0x0, 0xf3}
	data := []byte{}
	expected := []byte{0xb}
	callCode(t, client, user[0].PubKey.Address(), code, data, expected)

	// pass two ints as calldata, add, and return the result
	code = []byte{0x60, 0x0, 0x35, 0x60, 0x20, 0x35, 0x1, 0x60, 0x0, 0x52, 0x60,
		0x20, 0x60, 0x0, 0xf3}
	data = append(tm_common.LeftPadWord256([]byte{0x5}).Bytes(),
		tm_common.LeftPadWord256([]byte{0x6}).Bytes()...)
	expected = []byte{0xb}
	callCode(t, client, user[0].PubKey.Address(), code, data, expected)
}

func testCall(t *testing.T, typ string) {
	wsc := newWSClient(t)
	eid := txs.EventStringNewBlock()
	subscribe(t, wsc, eid)
	defer func() {
		unsubscribe(t, wsc, eid)
		wsc.Stop()
	}()

	client := clients[typ]

	// create the contract
	amt, gasLim, fee := int64(6969), int64(1000), int64(1000)
	code, _, _ := simpleContract()
	tx := makeDefaultCallTx(t, typ, nil, code, amt, gasLim, fee)
	receipt, err := broadcastTxAndWaitForBlock(t, typ, wsc, tx)
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
	callContract(t, client, user[0].PubKey.Address(), contractAddr, data, expected)
}

func testNameReg(t *testing.T, typ string) {
	wsc := newWSClient(t)

	txs.MinNameRegistrationPeriod = 1

	// register a new name, check if its there
	// since entries ought to be unique and these run against different clients, we append the typ
	name := "ye_old_domain_name_" + typ
	const data = "if not now, when"
	fee := int64(1000)
	numDesiredBlocks := int64(2)
	amt := fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*txs.NameBaseCost(name, data)

	tx := makeDefaultNameTx(t, typ, name, data, amt, fee)
	// verify the name by both using the event and by checking get_name
	subscribeAndWaitForNext(t, wsc, txs.EventStringNameReg(name),
		func() {
			broadcastTxAndWaitForBlock(t, typ, wsc, tx)
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

	entry := getNameRegEntry(t, typ, name)
	assert.Equal(t, data, entry.Data)
	assert.Equal(t, user[0].Address, entry.Owner)

	// update the data as the owner, make sure still there
	numDesiredBlocks = int64(5)
	const updatedData = "these are amongst the things I wish to bestow upon the youth of generations come: a safe supply of honey, and a better money. For what else shall they need"
	amt = fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*txs.NameBaseCost(name, updatedData)
	tx = makeDefaultNameTx(t, typ, name, updatedData, amt, fee)
	broadcastTxAndWaitForBlock(t, typ, wsc, tx)
	mempoolCount = 0
	entry = getNameRegEntry(t, typ, name)

	assert.Equal(t, updatedData, entry.Data)

	// try to update as non owner, should fail
	tx = txs.NewNameTxWithNonce(user[1].PubKey, name, "never mind", amt, fee,
		getNonce(t, typ, user[1].Address)+1)
	tx.Sign(chainID, user[1])

	_, err := broadcastTxAndWaitForBlock(t, typ, wsc, tx)
	assert.Error(t, err, "Expected error when updating someone else's unexpired"+
		" name registry entry")
	assert.Contains(t, err.Error(), "permission denied", "Error should be " +
			"permission denied")

	// Wait a couple of blocks to make sure name registration expires
	waitNBlocks(t, wsc, 3)

	//now the entry should be expired, so we can update as non owner
	const data2 = "this is not my beautiful house"
	tx = txs.NewNameTxWithNonce(user[1].PubKey, name, data2, amt, fee,
		getNonce(t, typ, user[1].Address)+1)
	tx.Sign(chainID, user[1])
	_, err = broadcastTxAndWaitForBlock(t, typ, wsc, tx)
	assert.NoError(t, err, "Should be able to update a previously expired name"+
		" registry entry as a different address")
	mempoolCount = 0
	entry = getNameRegEntry(t, typ, name)
	assert.Equal(t, data2, entry.Data)
	assert.Equal(t, user[1].Address, entry.Owner)
}

// ------ Test Utils -------

func asEventDataTx(t *testing.T, eventData txs.EventData) txs.EventDataTx {
	eventDataTx, ok := eventData.(txs.EventDataTx)
	if !ok {
		t.Fatalf("Expected eventData to be EventDataTx was %v", eventData)
	}
	return eventDataTx
}

func doNothing(_ string, _ txs.EventData) (bool, error) {
	// And ask waitForEvent to stop waiting
	return true, nil
}

func testSubscribe(t *testing.T) {
	var subId string
	wsc := newWSClient(t)
	subscribe(t, wsc, txs.EventStringNewBlock())

	timeout := time.NewTimer(timeoutSeconds * time.Second)
Subscribe:
	for {
		select {
		case <-timeout.C:
			t.Fatal("Timed out waiting for subscription result")

		case bs := <-wsc.ResultsCh:
			resultSubscribe, ok := readResult(t, bs).(*core_types.ResultSubscribe)
			if ok {
				assert.Equal(t, txs.EventStringNewBlock(), resultSubscribe.Event)
				subId = resultSubscribe.SubscriptionId
				break Subscribe
			}
		}
	}

	seenBlock := false
	timeout = time.NewTimer(timeoutSeconds * time.Second)
	for {
		select {
		case <-timeout.C:
			if !seenBlock {
				t.Fatal("Timed out without seeing a NewBlock event")
			}
			return

		case bs := <-wsc.ResultsCh:
			resultEvent, ok := readResult(t, bs).(*core_types.ResultEvent)
			if ok {
				_, ok := resultEvent.Data.(txs.EventDataNewBlock)
				if ok {
					if seenBlock {
						// There's a mild race here, but when we enter we've just seen a block
						// so we should be able to unsubscribe before we see another block
						t.Fatal("Continued to see NewBlock event after unsubscribing")
					} else {
						seenBlock = true
						unsubscribe(t, wsc, subId)
					}
				}
			}
		}
	}
}

func testBlockchainInfo(t *testing.T, typ string) {
	client := clients[typ]
	wsc := newWSClient(t)
	nBlocks := 4
	waitNBlocks(t, wsc, nBlocks)

	resp, err := edbcli.BlockchainInfo(client, 0, 0)
	if err != nil {
		t.Fatalf("Failed to get blockchain info: %v", err)
	}
	//TODO: [Silas] reintroduce this when Tendermint changes logic to fire
	// NewBlock after saving a block
	// see https://github.com/tendermint/tendermint/issues/273
	//assert.Equal(t, 4, resp.LastHeight, "Last height should be 4 after waiting for first 4 blocks")
	assert.True(t, nBlocks <= len(resp.BlockMetas),
		"Should see at least 4 BlockMetas after waiting for first 4 blocks")

	lastBlockHash := resp.BlockMetas[nBlocks-1].Hash
	for i := nBlocks - 2; i >= 0; i-- {
		assert.Equal(t, lastBlockHash, resp.BlockMetas[i].Header.LastBlockHash,
			"Blockchain should be a hash tree!")
		lastBlockHash = resp.BlockMetas[i].Hash
	}

	resp, err = edbcli.BlockchainInfo(client, 1, 2)
	if err != nil {
		t.Fatalf("Failed to get blockchain info: %v", err)
	}

	assert.Equal(t, 2, len(resp.BlockMetas),
		"Should see 2 BlockMetas after extracting 2 blocks")

	fmt.Printf("%v\n", resp)
}
