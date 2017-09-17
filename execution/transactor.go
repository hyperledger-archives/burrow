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
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/word"

	"github.com/hyperledger/burrow/blockchain"
	"github.com/tendermint/go-crypto"
)

// Transactor is the controller/middleware for the v0 RPC
type transactor struct {
	txMtx         *sync.Mutex
	blockchain    blockchain.Blockchain
	state         account.GetterAndStorageGetter
	eventEmitter  event.EventEmitter
	txBroadcaster func(tx txs.Tx) error
}

type Call struct {
	Return  string `json:"return"`
	GasUsed int64  `json:"gas_used"`
	// TODO ...
}

func newTransactor(blockchain blockchain.Blockchain,
	state account.GetterAndStorageGetter,
	eventEmitter event.EventEmitter,
	txBroadcaster func(tx txs.Tx) error) *transactor {
	return &transactor{
		blockchain:    blockchain,
		state:         state,
		eventEmitter:  eventEmitter,
		txBroadcaster: txBroadcaster,
	}
}

// Run a contract's code on an isolated and unpersisted state
// Cannot be used to create new contracts
func (trans *transactor) Call(fromAddress, toAddress account.Address, data []byte) (*Call, error) {
	// This was being run against CheckTx cache, need to understand the reasoning
	callee := trans.state.GetAccount(toAddress)
	if callee == nil {
		return nil, fmt.Errorf("account %s does not exist", toAddress)
	}
	caller := &account.ConcreteAccount{Address: fromAddress}
	txCache := NewTxCache(trans.state)
	params := vmParams(trans.blockchain)

	vmach := evm.NewVM(txCache, evm.DefaultDynamicMemoryProvider, params, caller.Address.Word256(), nil)
	vmach.SetFireable(trans.eventEmitter)

	gas := params.GasLimit
	ret, err := vmach.Call(caller, callee, callee.Code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	gasUsed := params.GasLimit - gas
	// here return bytes are hex encoded; on the sibling function
	// they are not
	return &Call{Return: hex.EncodeToString(ret), GasUsed: gasUsed}, nil
}

// Run the given code on an isolated and unpersisted state
// Cannot be used to create new contracts.
func (trans *transactor) CallCode(fromAddress account.Address, code, data []byte) (*Call, error) {
	// This was being run against CheckTx cache, need to understand the reasoning
	callee := &account.ConcreteAccount{Address: fromAddress}
	caller := &account.ConcreteAccount{Address: fromAddress}
	txCache := NewTxCache(trans.state)
	params := vmParams(trans.blockchain)

	vmach := evm.NewVM(txCache, evm.DefaultDynamicMemoryProvider, params, caller.Address.Word256(), nil)
	gas := params.GasLimit
	ret, err := vmach.Call(caller, callee, code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	gasUsed := params.GasLimit - gas
	// here return bytes are hex encoded; on the sibling function
	// they are not
	return &Call{Return: hex.EncodeToString(ret), GasUsed: gasUsed}, nil
}

// Broadcast a transaction.
func (trans *transactor) BroadcastTx(tx txs.Tx) (*txs.Receipt, error) {
	err := trans.txBroadcaster(tx)

	if err != nil {
		return nil, fmt.Errorf("Error broadcasting transaction: %v", err)
	}

	txHash := txs.TxHash(trans.blockchain.ChainID(), tx)
	var createsContract uint8
	var contractAddr account.Address
	// check if creates new contract
	if callTx, ok := tx.(*txs.CallTx); ok {
		if callTx.Address == account.ZeroAddress {
			createsContract = 1
			contractAddr = txs.NewContractAddress(callTx.Input.Address, callTx.Input.Sequence)
		}
	}
	return &txs.Receipt{txHash, createsContract, contractAddr}, nil
}

// Orders calls to BroadcastTx using lock (waits for response from core before releasing)
func (trans *transactor) Transact(privKey []byte, address account.Address, data []byte, gasLimit,
	fee int64) (*txs.Receipt, error) {
	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n", len(privKey))
	}
	trans.txMtx.Lock()
	defer trans.txMtx.Unlock()
	pa := account.GenPrivAccountFromPrivKeyBytes(privKey)
	acc := trans.state.GetAccount(pa.Address())
	sequence := int64(1)
	if acc != nil {
		sequence = acc.Sequence + 1
	}
	// TODO: [Silas] we should consider revising this method and removing fee, or
	// possibly adding an amount parameter. It is non-sensical to just be able to
	// set the fee. Our support of fees in general is questionable since at the
	// moment all we do is deduct the fee effectively leaking token. It is possible
	// someone may be using the sending of native token to payable functions but
	// they can be served by broadcasting a token.

	// We hard-code the amount to be equal to the fee which means the CallTx we
	// generate transfers 0 value, which is the most sensible default since in
	// recent solidity compilers the EVM generated will throw an error if value
	// is transferred to a non-payable function.
	txInput := &txs.TxInput{
		Address:  pa.Address(),
		Amount:   fee,
		Sequence: sequence,
		PubKey:   pa.PubKey(),
	}
	tx := &txs.CallTx{
		Input:    txInput,
		Address:  address,
		GasLimit: gasLimit,
		Fee:      fee,
		Data:     data,
	}

	// Got ourselves a tx.
	txS, errS := trans.SignTx(tx, []*account.ConcretePrivateAccount{pa.Unwrap()})
	if errS != nil {
		return nil, errS
	}
	return trans.BroadcastTx(txS)
}

func (trans *transactor) TransactAndHold(privKey []byte, address account.Address, data []byte, gasLimit,
	fee int64) (*evm.EventDataCall, error) {
	rec, tErr := trans.Transact(privKey, address, data, gasLimit, fee)
	if tErr != nil {
		return nil, tErr
	}
	var addr account.Address
	if rec.CreatesContract == 1 {
		addr = rec.ContractAddr
	} else {
		addr = address
	}
	// We want non-blocking on the first event received (but buffer the value),
	// after which we want to block (and then discard the value - see below)
	wc := make(chan *evm.EventDataCall, 1)
	subId := fmt.Sprintf("%X", rec.TxHash)
	trans.eventEmitter.Subscribe(subId, evm.EventStringAccCall(addr),
		func(evt evm.EventData) {
			eventDataCall := evt.(evm.EventDataCall)
			if bytes.Equal(eventDataCall.TxID, rec.TxHash) {
				// Beware the contract of go-events subscribe is that we must not be
				// blocking in an event callback when we try to unsubscribe!
				// We work around this by using a non-blocking send.
				select {
				// This is a non-blocking send, but since we are using a buffered
				// channel of size 1 we will always grab our first event even if we
				// haven't read from the channel at the time we receive the first event.
				case wc <- &eventDataCall:
				default:
				}
			}
		})

	timer := time.NewTimer(300 * time.Second)
	toChan := timer.C

	var ret *evm.EventDataCall
	var rErr error

	select {
	case <-toChan:
		rErr = fmt.Errorf("Transaction timed out. Hash: " + subId)
	case e := <-wc:
		timer.Stop()
		if e.Exception != "" {
			rErr = fmt.Errorf("error when transacting: " + e.Exception)
		} else {
			ret = e
		}
	}
	trans.eventEmitter.Unsubscribe(subId)
	return ret, rErr
}

func (trans *transactor) Send(privKey []byte, toAddress account.Address, amount int64) (*txs.Receipt, error) {
	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n",
			len(privKey))
	}

	pk := &[64]byte{}
	copy(pk[:], privKey)
	trans.txMtx.Lock()
	defer trans.txMtx.Unlock()
	pa := account.GenPrivAccountFromPrivKeyBytes(privKey)
	cache := trans.state
	acc := cache.GetAccount(pa.Address())
	sequence := int64(1)
	if acc != nil {
		sequence = acc.Sequence + 1
	}

	tx := txs.NewSendTx()

	txInput := &txs.TxInput{
		Address:  pa.Address(),
		Amount:   amount,
		Sequence: sequence,
		PubKey:   pa.PubKey(),
	}

	tx.Inputs = append(tx.Inputs, txInput)

	txOutput := &txs.TxOutput{Address: toAddress, Amount: amount}

	tx.Outputs = append(tx.Outputs, txOutput)

	// Got ourselves a tx.
	txS, errS := trans.SignTx(tx, []*account.ConcretePrivateAccount{pa.Unwrap()})
	if errS != nil {
		return nil, errS
	}
	return trans.BroadcastTx(txS)
}

func (trans *transactor) SendAndHold(privKey []byte, toAddress account.Address,
	amount int64) (*txs.Receipt, error) {
	rec, tErr := trans.Send(privKey, toAddress, amount)
	if tErr != nil {
		return nil, tErr
	}

	wc := make(chan *txs.SendTx)
	subId := fmt.Sprintf("%X", rec.TxHash)

	trans.eventEmitter.Subscribe(subId, evm.EventStringAccOutput(toAddress),
		func(evt evm.EventData) {
			eventDataTx := evt.(evm.EventDataTx)
			tx := eventDataTx.Tx.(*txs.SendTx)
			wc <- tx
		})

	timer := time.NewTimer(300 * time.Second)
	toChan := timer.C

	var rErr error

	pa := account.GenPrivAccountFromPrivKeyBytes(privKey)

	select {
	case <-toChan:
		rErr = fmt.Errorf("Transaction timed out. Hash: " + subId)
	case e := <-wc:
		if e.Inputs[0].Address == pa.Address() && e.Inputs[0].Amount == amount {
			timer.Stop()
			trans.eventEmitter.Unsubscribe(subId)
			return rec, rErr
		}
	}
	return nil, rErr
}

func (trans *transactor) TransactNameReg(privKey []byte, name, data string,
	amount, fee int64) (*txs.Receipt, error) {

	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n", len(privKey))
	}
	trans.txMtx.Lock()
	defer trans.txMtx.Unlock()
	pa := account.GenPrivAccountFromPrivKeyBytes(privKey)
	cache := trans.state // XXX: DON'T MUTATE THIS CACHE (used internally for CheckTx)
	acc := cache.GetAccount(pa.Address())
	sequence := int64(1)
	if acc == nil {
		sequence = acc.Sequence + 1
	}
	tx := txs.NewNameTxWithNonce(pa.PubKey(), name, data, amount, fee, sequence)
	// Got ourselves a tx.
	txS, errS := trans.SignTx(tx, []*account.ConcretePrivateAccount{pa.Unwrap()})
	if errS != nil {
		return nil, errS
	}
	return trans.BroadcastTx(txS)
}

// Sign a transaction
func (trans *transactor) SignTx(tx txs.Tx, privAccounts []*account.ConcretePrivateAccount) (txs.Tx, error) {
	// more checks?

	for i, privAccount := range privAccounts {
		if privAccount == nil || privAccount.PrivKey.Unwrap() == nil {
			return nil, fmt.Errorf("Invalid (empty) privAccount @%v", i)
		}
	}
	switch tx.(type) {
	case *txs.NameTx:
		nameTx := tx.(*txs.NameTx)
		nameTx.Input.PubKey = privAccounts[0].PubKey
		nameTx.Input.Signature = privAccounts[0].Sign(trans.blockchain.ChainID(), nameTx)
	case *txs.SendTx:
		sendTx := tx.(*txs.SendTx)
		for i, input := range sendTx.Inputs {
			input.PubKey = privAccounts[i].PubKey
			input.Signature = privAccounts[i].Sign(trans.blockchain.ChainID(), sendTx)
		}
	case *txs.CallTx:
		callTx := tx.(*txs.CallTx)
		callTx.Input.PubKey = privAccounts[0].PubKey
		callTx.Input.Signature = privAccounts[0].Sign(trans.blockchain.ChainID(), callTx)
	case *txs.BondTx:
		bondTx := tx.(*txs.BondTx)
		// the first privaccount corresponds to the BondTx pub key.
		// the rest to the inputs
		bondTx.Signature = privAccounts[0].Sign(trans.blockchain.ChainID(), bondTx).
			Unwrap().(crypto.SignatureEd25519)
		for i, input := range bondTx.Inputs {
			input.PubKey = privAccounts[i+1].PubKey
			input.Signature = privAccounts[i+1].Sign(trans.blockchain.ChainID(), bondTx)
		}
	case *txs.UnbondTx:
		unbondTx := tx.(*txs.UnbondTx)
		unbondTx.Signature = privAccounts[0].Sign(trans.blockchain.ChainID(), unbondTx).
			Unwrap().(crypto.SignatureEd25519)
	case *txs.RebondTx:
		rebondTx := tx.(*txs.RebondTx)
		rebondTx.Signature = privAccounts[0].Sign(trans.blockchain.ChainID(), rebondTx).
			Unwrap().(crypto.SignatureEd25519)
	default:
		return nil, fmt.Errorf("Object is not a proper transaction: %v\n", tx)
	}
	return tx, nil
}

func vmParams(blockchain blockchain.Blockchain) evm.Params {
	return evm.Params{
		BlockHeight: int64(blockchain.LastBlockHeight()),
		BlockHash:   word.LeftPadWord256(blockchain.LastBlockHash()),
		BlockTime:   blockchain.LastBlockTime().Unix(),
		GasLimit:    GasLimit,
	}
}
