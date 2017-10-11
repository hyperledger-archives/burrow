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

package execution

import (
	"bytes"
	"encoding/hex"
	"testing"

	"fmt"

	"github.com/hyperledger/burrow/execution/evm/sha3"

	"time"

	acm "github.com/hyperledger/burrow/account"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/word"
	"github.com/stretchr/testify/assert"
	tm_types "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tmlibs/db"
)

func init() {
	evm.SetDebug(true)
}

var deterministicGenesis = genesis.NewDeterministicGenesis(34059836243380576)
var testGenesisDoc, testPrivAccounts, _ = deterministicGenesis.
	GenesisDoc(3, true, 1000, 1, true, 1000)
var testChainID = testGenesisDoc.ChainID()

func execTxWithStateAndBlockchain(state *State, tip bcm.Tip, tx txs.Tx) error {
	exe := newExecutor(true, state, testChainID, tip, event.NewNoOpFireable(), logger)
	if err := exe.Execute(tx); err != nil {
		return err
	} else {
		exe.blockCache.Sync()
		return nil
	}
}

func execTxWithState(state *State, tx txs.Tx) error {
	return execTxWithStateAndBlockchain(state, bcm.NewBlockchain(testGenesisDoc), tx)
}

func commitNewBlock(state *State, blockchain bcm.MutableBlockchain) {
	blockchain.CommitBlock(blockchain.LastBlockTime().Add(time.Second), sha3.Sha3(blockchain.LastBlockHash()),
		state.Hash())
}

func execTxWithStateNewBlock(state *State, blockchain bcm.MutableBlockchain, tx txs.Tx) error {
	if err := execTxWithStateAndBlockchain(state, blockchain, tx); err != nil {
		return err
	}
	commitNewBlock(state, blockchain)
	return nil
}

func makeGenesisState(numAccounts int, randBalance bool, minBalance uint64, numValidators int, randBonded bool,
	minBonded int64) (*State, []acm.PrivateAccount, []*tm_types.PrivValidatorFS) {
	testGenesisDoc, privAccounts, privValidators := deterministicGenesis.GenesisDoc(numAccounts, randBalance, minBalance,
		numValidators, randBonded, minBonded)
	s0 := MakeGenesisState(dbm.NewMemDB(), testGenesisDoc)
	s0.Save()
	return s0, privAccounts, privValidators
}

func getAccount(state acm.Getter, address acm.Address) acm.MutableAccount {
	acc, _ := acm.GetMutableAccount(state, address)
	return acc
}

func addressPtr(account acm.Account) *acm.Address {
	if account == nil {
		return nil
	}
	accountAddresss := account.Address()
	return &accountAddresss
}

// Tests

func TestCopyState(t *testing.T) {
	// Generate a random state
	s0, privAccounts, _ := makeGenesisState(10, true, 1000, 5, true, 1000)
	s0Hash := s0.Hash()
	if len(s0Hash) == 0 {
		t.Error("Expected state hash")
	}

	// Check hash of copy
	s0Copy := s0.Copy()
	assert.Equal(t, s0Hash, s0Copy.Hash(), "Expected state copy hash to be the same")
	assert.Equal(t, s0Copy.Copy().Hash(), s0Copy.Hash(), "Expected COPY COPY COPY the same")

	// Mutate the original; hash should change.
	acc0Address := privAccounts[0].Address()
	acc := getAccount(s0, acc0Address)
	acc.AddToBalance(1)

	// The account balance shouldn't have changed yet.
	if getAccount(s0, acc0Address).Balance() == acc.Balance() {
		t.Error("Account balance changed unexpectedly")
	}

	// Setting, however, should change the balance.
	s0.UpdateAccount(acc)
	if getAccount(s0, acc0Address).Balance() != acc.Balance() {
		t.Error("Account balance wasn't set")
	}

	// Now that the state changed, the hash should change too.
	if bytes.Equal(s0Hash, s0.Hash()) {
		t.Error("Expected state hash to have changed")
	}

	// The s0Copy shouldn't have changed though.
	if !bytes.Equal(s0Hash, s0Copy.Hash()) {
		t.Error("Expected state copy hash to have not changed")
	}
}

/*
func makeBlock(t *testing.T, state *State, validation *tmtypes.Commit, txs []txs.Tx) *tmtypes.Block {
	if validation == nil {
		validation = &tmtypes.Commit{}
	}
	block := &tmtypes.Block{
		Header: &tmtypes.Header{
			testChainID:        testChainID,
			Height:         blockchain.LastBlockHeight() + 1,
			Time:           state.lastBlockTime.Add(time.Minute),
			NumTxs:         len(txs),
			lastBlockAppHash:  state.lastBlockAppHash,
			LastBlockParts: state.LastBlockParts,
			AppHash:        nil,
		},
		LastCommit: validation,
		Data: &tmtypes.Data{
			Txs: txs,
		},
	}
	block.FillHeader()

	// Fill in block StateHash
	err := state.ComputeBlockStateHash(block)
	if err != nil {
		t.Error("Error appending initial block:", err)
	}
	if len(block.Header.StateHash) == 0 {
		t.Error("Expected StateHash but got nothing.")
	}

	return block
}

func TestGenesisSaveLoad(t *testing.T) {

	// Generate a state, save & load it.
	s0, _, _ := makeGenesisState(10, true, 1000, 5, true, 1000)

	// Make complete block and blockParts
	block := makeBlock(t, s0, nil, nil)
	blockParts := block.MakePartSet()

	// Now append the block to s0.
	err := ExecBlock(s0, block, blockParts.Header())
	if err != nil {
		t.Error("Error appending initial block:", err)
	}

	// Save s0
	s0.Save()

	// Sanity check s0
	//s0.db.(*dbm.MemDB).Print()
	if s0.BondedValidators.TotalVotingPower() == 0 {
		t.Error("s0 BondedValidators TotalVotingPower should not be 0")
	}
	if s0.lastBlockHeight != 1 {
		t.Error("s0 lastBlockHeight should be 1, got", s0.lastBlockHeight)
	}

	// Load s1
	s1 := LoadState(s0.db)

	// Compare height & blockHash
	if s0.lastBlockHeight != s1.lastBlockHeight {
		t.Error("lastBlockHeight mismatch")
	}
	if !bytes.Equal(s0.lastBlockAppHash, s1.lastBlockAppHash) {
		t.Error("lastBlockAppHash mismatch")
	}

	// Compare state merkle trees
	if s0.BondedValidators.Size() != s1.BondedValidators.Size() {
		t.Error("BondedValidators Size mismatch")
	}
	if s0.BondedValidators.TotalVotingPower() != s1.BondedValidators.TotalVotingPower() {
		t.Error("BondedValidators TotalVotingPower mismatch")
	}
	if !bytes.Equal(s0.BondedValidators.Hash(), s1.BondedValidators.Hash()) {
		t.Error("BondedValidators hash mismatch")
	}
	if s0.UnbondingValidators.Size() != s1.UnbondingValidators.Size() {
		t.Error("UnbondingValidators Size mismatch")
	}
	if s0.UnbondingValidators.TotalVotingPower() != s1.UnbondingValidators.TotalVotingPower() {
		t.Error("UnbondingValidators TotalVotingPower mismatch")
	}
	if !bytes.Equal(s0.UnbondingValidators.Hash(), s1.UnbondingValidators.Hash()) {
		t.Error("UnbondingValidators hash mismatch")
	}
	if !bytes.Equal(s0.accounts.Hash(), s1.accounts.Hash()) {
		t.Error("Accounts mismatch")
	}
	if !bytes.Equal(s0.validatorInfos.Hash(), s1.validatorInfos.Hash()) {
		t.Error("Accounts mismatch")
	}
}
*/

func TestTxSequence(t *testing.T) {

	state, privAccounts, _ := makeGenesisState(3, true, 1000, 1, true, 1000)
	acc0 := getAccount(state, privAccounts[0].Address())
	acc0PubKey := privAccounts[0].PubKey()
	acc1 := getAccount(state, privAccounts[1].Address())

	// Test a variety of sequence numbers for the tx.
	// The tx should only pass when i == 1.
	for i := uint64(0); i < 3; i++ {
		sequence := acc0.Sequence() + i
		tx := txs.NewSendTx()
		tx.AddInputWithNonce(acc0PubKey, 1, sequence)
		tx.AddOutput(acc1.Address(), 1)
		tx.Inputs[0].Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		stateCopy := state.Copy()
		err := execTxWithState(stateCopy, tx)
		if i == 1 {
			// Sequence is good.
			if err != nil {
				t.Errorf("Expected good sequence to pass: %v", err)
			}
			// Check acc.Sequence().
			newAcc0 := getAccount(stateCopy, acc0.Address())
			if newAcc0.Sequence() != sequence {
				t.Errorf("Expected account sequence to change to %v, got %v",
					sequence, newAcc0.Sequence())
			}
		} else {
			// Sequence is bad.
			if err == nil {
				t.Errorf("Expected bad sequence to fail")
			}
			// Check acc.Sequence(). (shouldn't have changed)
			newAcc0 := getAccount(stateCopy, acc0.Address())
			if newAcc0.Sequence() != acc0.Sequence() {
				t.Errorf("Expected account sequence to not change from %v, got %v",
					acc0.Sequence(), newAcc0.Sequence())
			}
		}
	}
}

func TestNameTxs(t *testing.T) {
	state := MakeGenesisState(dbm.NewMemDB(), testGenesisDoc)
	state.Save()

	txs.MinNameRegistrationPeriod = 5
	blockchain := bcm.NewBlockchain(testGenesisDoc)
	startingBlock := blockchain.LastBlockHeight()

	// try some bad names. these should all fail
	names := []string{"", "\n", "123#$%", "\x00", string([]byte{20, 40, 60, 80}),
		"baffledbythespectacleinallofthisyouseeehesaidwithouteyessurprised", "no spaces please"}
	data := "something about all this just doesn't feel right."
	fee := uint64(1000)
	numDesiredBlocks := uint64(5)
	for _, name := range names {
		amt := fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*
			txs.NameBaseCost(name, data)
		tx, _ := txs.NewNameTx(state, testPrivAccounts[0].PubKey(), name, data, amt, fee)
		tx.Sign(testChainID, testPrivAccounts[0])

		if err := execTxWithState(state, tx); err == nil {
			t.Fatalf("Expected invalid name error from %s", name)
		}
	}

	// try some bad data. these should all fail
	name := "hold_it_chum"
	datas := []string{"cold&warm", "!@#$%^&*()", "<<<>>>>", "because why would you ever need a ~ or a & or even a % in a json file? make your case and we'll talk"}
	for _, data := range datas {
		amt := fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*
			txs.NameBaseCost(name, data)
		tx, _ := txs.NewNameTx(state, testPrivAccounts[0].PubKey(), name, data, amt, fee)
		tx.Sign(testChainID, testPrivAccounts[0])

		if err := execTxWithState(state, tx); err == nil {
			t.Fatalf("Expected invalid data error from %s", data)
		}
	}

	validateEntry := func(t *testing.T, entry *NameRegEntry, name, data string, addr acm.Address, expires uint64) {

		if entry == nil {
			t.Fatalf("Could not find name %s", name)
		}
		if entry.Owner != addr {
			t.Fatalf("Wrong owner. Got %s expected %s", entry.Owner, addr)
		}
		if data != entry.Data {
			t.Fatalf("Wrong data. Got %s expected %s", entry.Data, data)
		}
		if name != entry.Name {
			t.Fatalf("Wrong name. Got %s expected %s", entry.Name, name)
		}
		if expires != entry.Expires {
			t.Fatalf("Wrong expiry. Got %d, expected %d", entry.Expires, expires)
		}
	}

	// try a good one, check data, owner, expiry
	name = "@looking_good/karaoke_bar.broadband"
	data = "on this side of neptune there are 1234567890 people: first is OMNIVORE+-3. Or is it. Ok this is pretty restrictive. No exclamations :(. Faces tho :')"
	amt := fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*txs.NameBaseCost(name, data)
	tx, _ := txs.NewNameTx(state, testPrivAccounts[0].PubKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[0])
	if err := execTxWithState(state, tx); err != nil {
		t.Fatal(err)
	}
	entry := state.GetNameRegEntry(name)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), startingBlock+numDesiredBlocks)

	// fail to update it as non-owner, in same block
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PubKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithState(state, tx); err == nil {
		t.Fatal("Expected error")
	}

	// update it as owner, just to increase expiry, in same block
	// NOTE: we have to resend the data or it will clear it (is this what we want?)
	tx, _ = txs.NewNameTx(state, testPrivAccounts[0].PubKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[0])
	if err := execTxWithStateNewBlock(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry = state.GetNameRegEntry(name)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), startingBlock+numDesiredBlocks*2)

	// update it as owner, just to increase expiry, in next block
	tx, _ = txs.NewNameTx(state, testPrivAccounts[0].PubKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[0])
	if err := execTxWithStateNewBlock(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry = state.GetNameRegEntry(name)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), startingBlock+numDesiredBlocks*3)

	// fail to update it as non-owner
	// Fast forward
	for blockchain.Tip().LastBlockHeight() < entry.Expires-1 {
		commitNewBlock(state, blockchain)
	}
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PubKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithStateAndBlockchain(state, blockchain, tx); err == nil {
		t.Fatal("Expected error")
	}
	commitNewBlock(state, blockchain)

	// once expires, non-owner succeeds
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PubKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithStateAndBlockchain(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry = state.GetNameRegEntry(name)
	validateEntry(t, entry, name, data, testPrivAccounts[1].Address(), blockchain.LastBlockHeight()+numDesiredBlocks)

	// update it as new owner, with new data (longer), but keep the expiry!
	data = "In the beginning there was no thing, not even the beginning. It hadn't been here, no there, nor for that matter anywhere, not especially because it had not to even exist, let alone to not. Nothing especially odd about that."
	oldCredit := amt - fee
	numDesiredBlocks = 10
	amt = fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier* txs.NameBaseCost(name, data) - oldCredit
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PubKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithStateAndBlockchain(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry = state.GetNameRegEntry(name)
	validateEntry(t, entry, name, data, testPrivAccounts[1].Address(), blockchain.LastBlockHeight()+numDesiredBlocks)

	// test removal
	amt = fee
	data = ""
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PubKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithStateNewBlock(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry = state.GetNameRegEntry(name)
	if entry != nil {
		t.Fatal("Expected removed entry to be nil")
	}

	// create entry by key0,
	// test removal by key1 after expiry
	name = "looking_good/karaoke_bar"
	data = "some data"
	amt = fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*txs.NameBaseCost(name, data)
	tx, _ = txs.NewNameTx(state, testPrivAccounts[0].PubKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[0])
	if err := execTxWithStateAndBlockchain(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry = state.GetNameRegEntry(name)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), blockchain.LastBlockHeight()+numDesiredBlocks)
	// Fast forward
	for blockchain.Tip().LastBlockHeight() < entry.Expires {
		commitNewBlock(state, blockchain)
	}

	amt = fee
	data = ""
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PubKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithStateNewBlock(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry = state.GetNameRegEntry(name)
	if entry != nil {
		t.Fatal("Expected removed entry to be nil")
	}
}

// Test creating a contract from futher down the call stack
/*
contract Factory {
    address a;
    function create() returns (address){
        a = new PreFactory();
        return a;
    }
}

contract PreFactory{
    address a;
    function create(Factory c) returns (address) {
    	a = c.create();
    	return a;
    }
}
*/

// run-time byte code for each of the above
var preFactoryCode, _ = hex.DecodeString("60606040526000357C0100000000000000000000000000000000000000000000000000000000900480639ED933181461003957610037565B005B61004F600480803590602001909190505061007B565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B60008173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1663EFC81A8C604051817C01000000000000000000000000000000000000000000000000000000000281526004018090506020604051808303816000876161DA5A03F1156100025750505060405180519060200150600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905061013C565B91905056")
var factoryCode, _ = hex.DecodeString("60606040526000357C010000000000000000000000000000000000000000000000000000000090048063EFC81A8C146037576035565B005B60426004805050606E565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B6000604051610153806100E0833901809050604051809103906000F0600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905060DD565B90566060604052610141806100126000396000F360606040526000357C0100000000000000000000000000000000000000000000000000000000900480639ED933181461003957610037565B005B61004F600480803590602001909190505061007B565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B60008173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1663EFC81A8C604051817C01000000000000000000000000000000000000000000000000000000000281526004018090506020604051808303816000876161DA5A03F1156100025750505060405180519060200150600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905061013C565B91905056")
var createData, _ = hex.DecodeString("9ed93318")

func TestCreates(t *testing.T) {
	//evm.SetDebug(true)
	state, privAccounts, _ := makeGenesisState(3, true, 1000, 1, true, 1000)

	//val0 := state.GetValidatorInfo(privValidators[0].Address())
	acc0 := getAccount(state, privAccounts[0].Address())
	acc0PubKey := privAccounts[0].PubKey()
	acc1 := getAccount(state, privAccounts[1].Address())
	acc2 := getAccount(state, privAccounts[2].Address())

	state = state.Copy()
	newAcc1 := getAccount(state, acc1.Address())
	newAcc1.SetCode(preFactoryCode)
	newAcc2 := getAccount(state, acc2.Address())
	newAcc2.SetCode(factoryCode)

	state.UpdateAccount(newAcc1)
	state.UpdateAccount(newAcc2)

	createData = append(createData, acc2.Address().Word256().Bytes()...)

	// call the pre-factory, triggering the factory to run a create
	tx := &txs.CallTx{
		Input: &txs.TxInput{
			Address:  acc0.Address(),
			Amount:   1,
			Sequence: acc0.Sequence() + 1,
			PubKey:   acc0PubKey,
		},
		Address:  addressPtr(acc1),
		GasLimit: 10000,
		Data:     createData,
	}

	tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
	err := execTxWithState(state, tx)
	if err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

	acc1 = getAccount(state, acc1.Address())
	storage := state.LoadStorage(acc1.StorageRoot())
	_, firstCreatedAddress, _ := storage.Get(word.LeftPadBytes([]byte{0}, 32))

	acc0 = getAccount(state, acc0.Address())
	// call the pre-factory, triggering the factory to run a create
	tx = &txs.CallTx{
		Input: &txs.TxInput{
			Address:  acc0.Address(),
			Amount:   1,
			Sequence: acc0.Sequence() + 1,
			PubKey:   acc0PubKey,
		},
		Address:  addressPtr(acc1),
		GasLimit: 100000,
		Data:     createData,
	}

	tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
	err = execTxWithState(state, tx)
	if err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

	acc1 = getAccount(state, acc1.Address())
	storage = state.LoadStorage(acc1.StorageRoot())
	_, secondCreatedAddress, _ := storage.Get(word.LeftPadBytes([]byte{0}, 32))

	if bytes.Equal(firstCreatedAddress, secondCreatedAddress) {
		t.Errorf("Multiple contracts created with the same address!")
	}
}

/*
contract Caller {
    function send(address x){
        x.send(msg.value);
    }
}
*/
var callerCode, _ = hex.DecodeString("60606040526000357c0100000000000000000000000000000000000000000000000000000000900480633e58c58c146037576035565b005b604b6004808035906020019091905050604d565b005b8073ffffffffffffffffffffffffffffffffffffffff16600034604051809050600060405180830381858888f19350505050505b5056")
var sendData, _ = hex.DecodeString("3e58c58c")

func TestContractSend(t *testing.T) {
	state, privAccounts, _ := makeGenesisState(3, true, 1000, 1, true, 1000)

	//val0 := state.GetValidatorInfo(privValidators[0].Address())
	acc0 := getAccount(state, privAccounts[0].Address())
	acc0PubKey := privAccounts[0].PubKey()
	acc1 := getAccount(state, privAccounts[1].Address())
	acc2 := getAccount(state, privAccounts[2].Address())

	state = state.Copy()
	newAcc1 := getAccount(state, acc1.Address())
	newAcc1.SetCode(callerCode)
	state.UpdateAccount(newAcc1)

	sendData = append(sendData, acc2.Address().Word256().Bytes()...)
	sendAmt := uint64(10)
	acc2Balance := acc2.Balance()

	// call the contract, triggering the send
	tx := &txs.CallTx{
		Input: &txs.TxInput{
			Address:  acc0.Address(),
			Amount:   sendAmt,
			Sequence: acc0.Sequence() + 1,
			PubKey:   acc0PubKey,
		},
		Address:  addressPtr(acc1),
		GasLimit: 1000,
		Data:     sendData,
	}

	tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
	err := execTxWithState(state, tx)
	if err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

	acc2 = getAccount(state, acc2.Address())
	if acc2.Balance() != sendAmt+acc2Balance {
		t.Errorf("Value transfer from contract failed! Got %d, expected %d", acc2.Balance(), sendAmt+acc2Balance)
	}
}
func TestMerklePanic(t *testing.T) {
	state, privAccounts, _ := makeGenesisState(3, true, 1000,
		1, true, 1000)

	//val0 := state.GetValidatorInfo(privValidators[0].Address())
	acc0 := getAccount(state, privAccounts[0].Address())
	acc0PubKey := privAccounts[0].PubKey()
	acc1 := getAccount(state, privAccounts[1].Address())

	state.Save()
	// SendTx.
	{
		stateSendTx := state.Copy()
		tx := &txs.SendTx{
			Inputs: []*txs.TxInput{
				{
					Address:  acc0.Address(),
					Amount:   1,
					Sequence: acc0.Sequence() + 1,
					PubKey:   acc0PubKey,
				},
			},
			Outputs: []*txs.TxOutput{
				{
					Address: acc1.Address(),
					Amount:  1,
				},
			},
		}

		tx.Inputs[0].Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		err := execTxWithState(stateSendTx, tx)
		if err != nil {
			t.Errorf("Got error in executing send transaction, %v", err)
		}
		// uncomment for panic fun!
		//stateSendTx.Save()
	}

	// CallTx. Just runs through it and checks the transfer. See vm, rpc tests for more
	{
		stateCallTx := state.Copy()
		newAcc1 := getAccount(stateCallTx, acc1.Address())
		newAcc1.SetCode([]byte{0x60})
		stateCallTx.UpdateAccount(newAcc1)
		tx := &txs.CallTx{
			Input: &txs.TxInput{
				Address:  acc0.Address(),
				Amount:   1,
				Sequence: acc0.Sequence() + 1,
				PubKey:   acc0PubKey,
			},
			Address:  addressPtr(acc1),
			GasLimit: 10,
		}

		tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		err := execTxWithState(stateCallTx, tx)
		if err != nil {
			t.Errorf("Got error in executing call transaction, %v", err)
		}
	}
	state.Save()
	trygetacc0 := getAccount(state, privAccounts[0].Address())
	fmt.Println(trygetacc0.Address())
}

// TODO: test overflows.
// TODO: test for unbonding validators.
func TestTxs(t *testing.T) {
	state, privAccounts, _ := makeGenesisState(3, true, 1000, 1, true, 1000)

	//val0 := state.GetValidatorInfo(privValidators[0].Address())
	acc0 := getAccount(state, privAccounts[0].Address())
	acc0PubKey := privAccounts[0].PubKey()
	acc1 := getAccount(state, privAccounts[1].Address())

	// SendTx.
	{
		stateSendTx := state.Copy()
		tx := &txs.SendTx{
			Inputs: []*txs.TxInput{
				{
					Address:  acc0.Address(),
					Amount:   1,
					Sequence: acc0.Sequence() + 1,
					PubKey:   acc0PubKey,
				},
			},
			Outputs: []*txs.TxOutput{
				{
					Address: acc1.Address(),
					Amount:  1,
				},
			},
		}

		tx.Inputs[0].Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		err := execTxWithState(stateSendTx, tx)
		if err != nil {
			t.Errorf("Got error in executing send transaction, %v", err)
		}
		newAcc0 := getAccount(stateSendTx, acc0.Address())
		if acc0.Balance()-1 != newAcc0.Balance() {
			t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
				acc0.Balance()-1, newAcc0.Balance())
		}
		newAcc1 := getAccount(stateSendTx, acc1.Address())
		if acc1.Balance()+1 != newAcc1.Balance() {
			t.Errorf("Unexpected newAcc1 balance. Expected %v, got %v",
				acc1.Balance()+1, newAcc1.Balance())
		}
	}

	// CallTx. Just runs through it and checks the transfer. See vm, rpc tests for more
	{
		stateCallTx := state.Copy()
		newAcc1 := getAccount(stateCallTx, acc1.Address())
		newAcc1.SetCode([]byte{0x60})
		stateCallTx.UpdateAccount(newAcc1)
		tx := &txs.CallTx{
			Input: &txs.TxInput{
				Address:  acc0.Address(),
				Amount:   1,
				Sequence: acc0.Sequence() + 1,
				PubKey:   acc0PubKey,
			},
			Address:  addressPtr(acc1),
			GasLimit: 10,
		}

		tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		err := execTxWithState(stateCallTx, tx)
		if err != nil {
			t.Errorf("Got error in executing call transaction, %v", err)
		}
		newAcc0 := getAccount(stateCallTx, acc0.Address())
		if acc0.Balance()-1 != newAcc0.Balance() {
			t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
				acc0.Balance()-1, newAcc0.Balance())
		}
		newAcc1 = getAccount(stateCallTx, acc1.Address())
		if acc1.Balance()+1 != newAcc1.Balance() {
			t.Errorf("Unexpected newAcc1 balance. Expected %v, got %v",
				acc1.Balance()+1, newAcc1.Balance())
		}
	}
	trygetacc0 := getAccount(state, privAccounts[0].Address())
	fmt.Println(trygetacc0.Address())

	// NameTx.
	{
		entryName := "satoshi"
		entryData := `
A  purely   peer-to-peer   version   of   electronic   cash   would   allow   online
payments  to  be  sent   directly  from  one  party  to  another  without   going  through  a
financial institution.   Digital signatures provide part of the solution, but the main
benefits are lost if a trusted third party is still required to prevent double-spending.
We propose a solution to the double-spending problem using a peer-to-peer network.
The   network   timestamps   transactions  by  hashing   them   into   an   ongoing   chain   of
hash-based proof-of-work, forming a record that cannot be changed without redoing
the proof-of-work.   The longest chain not only serves as proof of the sequence of
events witnessed, but proof that it came from the largest pool of CPU power.   As
long as a majority of CPU power is controlled by nodes that are not cooperating to
attack the network, they'll generate the longest chain and outpace attackers.   The
network itself requires minimal structure.   Messages are broadcast on a best effort
basis,   and   nodes   can   leave  and   rejoin   the  network   at  will,  accepting   the   longest
proof-of-work chain as proof of what happened while they were gone `
		entryAmount := uint64(10000)

		stateNameTx := state.Copy()
		tx := &txs.NameTx{
			Input: &txs.TxInput{
				Address:  acc0.Address(),
				Amount:   entryAmount,
				Sequence: acc0.Sequence() + 1,
				PubKey:   acc0PubKey,
			},
			Name: entryName,
			Data: entryData,
		}

		tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)

		err := execTxWithState(stateNameTx, tx)
		if err != nil {
			t.Errorf("Got error in executing call transaction, %v", err)
		}
		newAcc0 := getAccount(stateNameTx, acc0.Address())
		if acc0.Balance()-entryAmount != newAcc0.Balance() {
			t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
				acc0.Balance()-entryAmount, newAcc0.Balance())
		}
		entry := stateNameTx.GetNameRegEntry(entryName)
		if entry == nil {
			t.Errorf("Expected an entry but got nil")
		}
		if entry.Data != entryData {
			t.Errorf("Wrong data stored")
		}

		// test a bad string
		tx.Data = string([]byte{0, 1, 2, 3, 127, 128, 129, 200, 251})
		tx.Input.Sequence += 1
		tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		err = execTxWithState(stateNameTx, tx)
		if _, ok := err.(txs.ErrTxInvalidString); !ok {
			t.Errorf("Expected invalid string error. Got: %s", err.Error())
		}
	}

	// BondTx. TODO
	/*
		{
			state := state.Copy()
			tx := &txs.BondTx{
				PubKey: acc0PubKey.(crypto.PubKeyEd25519),
				Inputs: []*txs.TxInput{
					&txs.TxInput{
						Address:  acc0.Address(),
						Amount:   1,
						Sequence: acc0.Sequence() + 1,
						PubKey:   acc0PubKey,
					},
				},
				UnbondTo: []*txs.TxOutput{
					&txs.TxOutput{
						Address: acc0.Address(),
						Amount:  1,
					},
				},
			}
			tx.Signature = privAccounts[0] acm.ChainSign(testChainID, tx).(crypto.SignatureEd25519)
			tx.Inputs[0].Signature = privAccounts[0] acm.ChainSign(testChainID, tx)
			err := execTxWithState(state, tx)
			if err != nil {
				t.Errorf("Got error in executing bond transaction, %v", err)
			}
			newAcc0 := getAccount(state, acc0.Address())
			if newAcc0.Balance() != acc0.Balance()-1 {
				t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
					acc0.Balance()-1, newAcc0.Balance())
			}
			_, acc0Val := state.BondedValidators.GetByAddress(acc0.Address())
			if acc0Val == nil {
				t.Errorf("acc0Val not present")
			}
			if acc0Val.BondHeight != blockchain.LastBlockHeight()+1 {
				t.Errorf("Unexpected bond height. Expected %v, got %v",
					blockchain.LastBlockHeight(), acc0Val.BondHeight)
			}
			if acc0Val.VotingPower != 1 {
				t.Errorf("Unexpected voting power. Expected %v, got %v",
					acc0Val.VotingPower, acc0.Balance())
			}
			if acc0Val.Accum != 0 {
				t.Errorf("Unexpected accum. Expected 0, got %v",
					acc0Val.Accum)
			}
		} */

	// TODO UnbondTx.

}

func TestSelfDestruct(t *testing.T) {

	state, privAccounts, _ := makeGenesisState(3, true, 1000, 1, true, 1000)

	acc0 := getAccount(state, privAccounts[0].Address())
	acc0PubKey := privAccounts[0].PubKey()
	acc1 := getAccount(state, privAccounts[1].Address())
	acc2 := getAccount(state, privAccounts[2].Address())
	sendingAmount, refundedBalance, oldBalance := uint64(1), acc1.Balance(), acc2.Balance()

	newAcc1 := getAccount(state, acc1.Address())

	// store 0x1 at 0x1, push an address, then self-destruct:)
	contractCode := []byte{0x60, 0x01, 0x60, 0x01, 0x55, 0x73}
	contractCode = append(contractCode, acc2.Address().Bytes()...)
	contractCode = append(contractCode, 0xff)
	newAcc1.SetCode(contractCode)
	state.UpdateAccount(newAcc1)

	// send call tx with no data, cause self-destruct
	tx := txs.NewCallTxWithNonce(acc0PubKey, addressPtr(acc1), nil, sendingAmount, 1000, 0, acc0.Sequence()+1)
	tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)

	// we use cache instead of execTxWithState so we can run the tx twice
	exe := NewBatchCommitter(state, testChainID, bcm.NewBlockchain(testGenesisDoc), event.NewNoOpFireable(), logger)
	if err := exe.Execute(tx); err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

	// if we do it again, we won't get an error, but the self-destruct
	// shouldn't happen twice and the caller should lose fee
	tx.Input.Sequence += 1
	tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
	if err := exe.Execute(tx); err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

	// commit the block
	exe.Commit()

	// acc2 should receive the sent funds and the contracts balance
	newAcc2 := getAccount(state, acc2.Address())
	newBalance := sendingAmount + refundedBalance + oldBalance
	if newAcc2.Balance() != newBalance {
		t.Errorf("Unexpected newAcc2 balance. Expected %v, got %v",
			newAcc2.Balance(), newBalance)
	}
	newAcc1 = getAccount(state, acc1.Address())
	if newAcc1 != nil {
		t.Errorf("Expected account to be removed")
	}

}

/* TODO
func TestAddValidator(t *testing.T) {

	// Generate a state, save & load it.
	s0, privAccounts, privValidators := makeGenesisState(10, false, 1000, 1, false, 1000)

	// The first privAccount will become a validator
	acc0 := privAccounts[0]
	bondTx := &txs.BondTx{
		PubKey: acc0.PubKey.(account.PubKeyEd25519),
		Inputs: []*txs.TxInput{
			&txs.TxInput{
				Address:  acc0.Address(),
				Amount:   1000,
				Sequence: 1,
				PubKey:   acc0.PubKey,
			},
		},
		UnbondTo: []*txs.TxOutput{
			&txs.TxOutput{
				Address: acc0.Address(),
				Amount:  1000,
			},
		},
	}
	bondTx.Signature = acc0 acm.ChainSign(testChainID, bondTx).(account.SignatureEd25519)
	bondTx.Inputs[0].Signature = acc0 acm.ChainSign(testChainID, bondTx)

	// Make complete block and blockParts
	block0 := makeBlock(t, s0, nil, []txs.Tx{bondTx})
	block0Parts := block0.MakePartSet()

	// Sanity check
	if s0.BondedValidators.Size() != 1 {
		t.Error("Expected there to be 1 validators before bondTx")
	}

	// Now append the block to s0.
	err := ExecBlock(s0, block0, block0Parts.Header())
	if err != nil {
		t.Error("Error appending initial block:", err)
	}

	// Must save before further modification
	s0.Save()

	// Test new validator set
	if s0.BondedValidators.Size() != 2 {
		t.Error("Expected there to be 2 validators after bondTx")
	}

	// The validation for the next block should only require 1 signature
	// (the new validator wasn't active for block0)
	precommit0 := &txs.Vote{
		Height:           1,
		Round:            0,
		Type:             txs.VoteTypePrecommit,
		BlockHash:        block0.Hash(),
		BlockPartsHeader: block0Parts.Header(),
	}
	privValidators[0].SignVote(testChainID, precommit0)

	block1 := makeBlock(t, s0,
		&txs.Validation{
			Precommits: []*txs.Vote{
				precommit0,
			},
		}, nil,
	)
	block1Parts := block1.MakePartSet()
	err = ExecBlock(s0, block1, block1Parts.Header())
	if err != nil {
		t.Error("Error appending secondary block:", err)
	}
}
*/
