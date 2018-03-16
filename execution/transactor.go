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
	"context"
	"fmt"
	"sync"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	"github.com/hyperledger/burrow/event"
	exe_events "github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/evm"
	evm_events "github.com/hyperledger/burrow/execution/evm/events"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/txs"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
)

const BlockingTimeoutSeconds = 30

type Call struct {
	Return  []byte
	GasUsed uint64
}

type Transactor interface {
	Call(fromAddress, toAddress acm.Address, data []byte) (*Call, error)
	CallCode(fromAddress acm.Address, code, data []byte) (*Call, error)
	BroadcastTx(tx txs.Tx) (*txs.Receipt, error)
	BroadcastTxAsync(tx txs.Tx, callback func(res *abci_types.Response)) error
	Transact(privKey []byte, address *acm.Address, data []byte, gasLimit, fee uint64) (*txs.Receipt, error)
	TransactAndHold(privKey []byte, address *acm.Address, data []byte, gasLimit, fee uint64) (*evm_events.EventDataCall, error)
	Send(privKey []byte, toAddress acm.Address, amount uint64) (*txs.Receipt, error)
	SendAndHold(privKey []byte, toAddress acm.Address, amount uint64) (*txs.Receipt, error)
	TransactNameReg(privKey []byte, name, data string, amount, fee uint64) (*txs.Receipt, error)
	SignTx(tx txs.Tx, privAccounts []acm.PrivateAccount) (txs.Tx, error)
}

// Transactor is the controller/middleware for the v0 RPC
type transactor struct {
	txMtx            sync.Mutex
	blockchain       blockchain.Blockchain
	state            acm.StateIterable
	eventEmitter     event.Emitter
	broadcastTxAsync func(tx txs.Tx, callback func(res *abci_types.Response)) error
	logger           logging_types.InfoTraceLogger
}

var _ Transactor = &transactor{}

func NewTransactor(blockchain blockchain.Blockchain, state acm.StateIterable, eventEmitter event.Emitter,
	broadcastTxAsync func(tx txs.Tx, callback func(res *abci_types.Response)) error,
	logger logging_types.InfoTraceLogger) *transactor {

	return &transactor{
		blockchain:       blockchain,
		state:            state,
		eventEmitter:     eventEmitter,
		broadcastTxAsync: broadcastTxAsync,
		logger:           logger.With(structure.ComponentKey, "Transactor"),
	}
}

// Run a contract's code on an isolated and unpersisted state
// Cannot be used to create new contracts
func (trans *transactor) Call(fromAddress, toAddress acm.Address, data []byte) (*Call, error) {
	if evm.RegisteredNativeContract(toAddress.Word256()) {
		return nil, fmt.Errorf("attempt to call native contract at address "+
			"%X, but native contracts can not be called directly. Use a deployed "+
			"contract that calls the native function instead", toAddress)
	}
	// This was being run against CheckTx cache, need to understand the reasoning
	callee, err := acm.GetMutableAccount(trans.state, toAddress)
	if err != nil {
		return nil, err
	}
	if callee == nil {
		return nil, fmt.Errorf("account %s does not exist", toAddress)
	}
	caller := acm.ConcreteAccount{Address: fromAddress}.MutableAccount()
	txCache := acm.NewStateCache(trans.state)
	params := vmParams(trans.blockchain)

	vmach := evm.NewVM(txCache, evm.DefaultDynamicMemoryProvider, params, caller.Address(), nil,
		logging.WithScope(trans.logger, "Call"))
	vmach.SetPublisher(trans.eventEmitter)

	gas := params.GasLimit
	ret, err := vmach.Call(caller, callee, callee.Code(), data, 0, &gas)
	if err != nil {
		return nil, err
	}
	gasUsed := params.GasLimit - gas
	return &Call{Return: ret, GasUsed: gasUsed}, nil
}

// Run the given code on an isolated and unpersisted state
// Cannot be used to create new contracts.
func (trans *transactor) CallCode(fromAddress acm.Address, code, data []byte) (*Call, error) {
	// This was being run against CheckTx cache, need to understand the reasoning
	callee := acm.ConcreteAccount{Address: fromAddress}.MutableAccount()
	caller := acm.ConcreteAccount{Address: fromAddress}.MutableAccount()
	txCache := acm.NewStateCache(trans.state)
	params := vmParams(trans.blockchain)

	vmach := evm.NewVM(txCache, evm.DefaultDynamicMemoryProvider, params, caller.Address(), nil,
		logging.WithScope(trans.logger, "CallCode"))
	gas := params.GasLimit
	ret, err := vmach.Call(caller, callee, code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	gasUsed := params.GasLimit - gas
	return &Call{Return: ret, GasUsed: gasUsed}, nil
}

func (trans *transactor) BroadcastTxAsync(tx txs.Tx, callback func(res *abci_types.Response)) error {
	return trans.broadcastTxAsync(tx, callback)
}

// Broadcast a transaction.
func (trans *transactor) BroadcastTx(tx txs.Tx) (*txs.Receipt, error) {
	responseCh := make(chan *abci_types.Response, 1)
	err := trans.BroadcastTxAsync(tx, func(res *abci_types.Response) {
		responseCh <- res
	})

	if err != nil {
		return nil, err
	}
	response := <-responseCh
	checkTxResponse := response.GetCheckTx()
	if checkTxResponse == nil {
		return nil, fmt.Errorf("application did not return CheckTx response")
	}

	switch checkTxResponse.Code {
	case codes.TxExecutionSuccessCode:
		receipt := new(txs.Receipt)
		err := wire.ReadBinaryBytes(checkTxResponse.Data, receipt)
		if err != nil {
			return nil, fmt.Errorf("could not deserialise transaction receipt: %s", err)
		}
		return receipt, nil
	default:
		return nil, fmt.Errorf("error returned by Tendermint in BroadcastTxSync "+
			"ABCI code: %v, ABCI log: %v", checkTxResponse.Code, checkTxResponse.Log)
	}
}

// Orders calls to BroadcastTx using lock (waits for response from core before releasing)
func (trans *transactor) Transact(privKey []byte, address *acm.Address, data []byte, gasLimit,
	fee uint64) (*txs.Receipt, error) {

	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n", len(privKey))
	}
	trans.txMtx.Lock()
	defer trans.txMtx.Unlock()
	pa, err := acm.GeneratePrivateAccountFromPrivateKeyBytes(privKey)
	if err != nil {
		return nil, err
	}
	// [Silas] This is puzzling, if the account doesn't exist the CallTx will fail, so what's the point in this?
	acc, err := trans.state.GetAccount(pa.Address())
	if err != nil {
		return nil, err
	}
	sequence := uint64(1)
	if acc != nil {
		sequence = acc.Sequence() + uint64(1)
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
		Address:   pa.Address(),
		Amount:    fee,
		Sequence:  sequence,
		PublicKey: pa.PublicKey(),
	}
	tx := &txs.CallTx{
		Input:    txInput,
		Address:  address,
		GasLimit: gasLimit,
		Fee:      fee,
		Data:     data,
	}

	// Got ourselves a tx.
	txS, errS := trans.SignTx(tx, []acm.PrivateAccount{pa})
	if errS != nil {
		return nil, errS
	}
	return trans.BroadcastTx(txS)
}

func (trans *transactor) TransactAndHold(privKey []byte, address *acm.Address, data []byte, gasLimit,
	fee uint64) (*evm_events.EventDataCall, error) {

	receipt, err := trans.Transact(privKey, address, data, gasLimit, fee)
	if err != nil {
		return nil, err
	}

	// We want non-blocking on the first event received (but buffer the value),
	// after which we want to block (and then discard the value - see below)
	wc := make(chan *evm_events.EventDataCall, 1)

	subID, err := event.GenerateSubscriptionID()
	if err != nil {
		return nil, err
	}

	err = evm_events.SubscribeAccountCall(context.Background(), trans.eventEmitter, subID, receipt.ContractAddress,
		receipt.TxHash, wc)
	if err != nil {
		return nil, err
	}
	// Will clean up callback goroutine and subscription in pubsub
	defer trans.eventEmitter.UnsubscribeAll(context.Background(), subID)

	timer := time.NewTimer(BlockingTimeoutSeconds * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil, fmt.Errorf("transaction timed out TxHash: %X", receipt.TxHash)
	case eventDataCall := <-wc:
		if eventDataCall.Exception != "" {
			return nil, fmt.Errorf("error when transacting: " + eventDataCall.Exception)
		} else {
			return eventDataCall, nil
		}
	}
}

func (trans *transactor) Send(privKey []byte, toAddress acm.Address, amount uint64) (*txs.Receipt, error) {
	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n",
			len(privKey))
	}

	pk := &[64]byte{}
	copy(pk[:], privKey)
	trans.txMtx.Lock()
	defer trans.txMtx.Unlock()
	pa, err := acm.GeneratePrivateAccountFromPrivateKeyBytes(privKey)
	if err != nil {
		return nil, err
	}
	cache := trans.state
	acc, err := cache.GetAccount(pa.Address())
	if err != nil {
		return nil, err
	}
	sequence := uint64(1)
	if acc != nil {
		sequence = acc.Sequence() + uint64(1)
	}

	tx := txs.NewSendTx()

	txInput := &txs.TxInput{
		Address:   pa.Address(),
		Amount:    amount,
		Sequence:  sequence,
		PublicKey: pa.PublicKey(),
	}

	tx.Inputs = append(tx.Inputs, txInput)

	txOutput := &txs.TxOutput{Address: toAddress, Amount: amount}

	tx.Outputs = append(tx.Outputs, txOutput)

	// Got ourselves a tx.
	txS, errS := trans.SignTx(tx, []acm.PrivateAccount{pa})
	if errS != nil {
		return nil, errS
	}
	return trans.BroadcastTx(txS)
}

func (trans *transactor) SendAndHold(privKey []byte, toAddress acm.Address, amount uint64) (*txs.Receipt, error) {
	receipt, err := trans.Send(privKey, toAddress, amount)
	if err != nil {
		return nil, err
	}

	wc := make(chan *txs.SendTx)

	subID, err := event.GenerateSubscriptionID()
	if err != nil {
		return nil, err
	}

	err = exe_events.SubscribeAccountOutputSendTx(context.Background(), trans.eventEmitter, subID, toAddress,
		receipt.TxHash, wc)
	if err != nil {
		return nil, err
	}
	defer trans.eventEmitter.UnsubscribeAll(context.Background(), subID)

	timer := time.NewTimer(BlockingTimeoutSeconds * time.Second)
	defer timer.Stop()

	pa, err := acm.GeneratePrivateAccountFromPrivateKeyBytes(privKey)
	if err != nil {
		return nil, err
	}

	select {
	case <-timer.C:
		return nil, fmt.Errorf("transaction timed out TxHash: %X", receipt.TxHash)
	case sendTx := <-wc:
		// This is a double check - we subscribed to this tx's hash so something has gone wrong if the amounts don't match
		if sendTx.Inputs[0].Address == pa.Address() && sendTx.Inputs[0].Amount == amount {
			return receipt, nil
		}
		return nil, fmt.Errorf("received SendTx but hash doesn't seem to match what we subscribed to, "+
			"received SendTx: %v which does not match receipt on sending: %v", sendTx, receipt)
	}
}

func (trans *transactor) TransactNameReg(privKey []byte, name, data string, amount, fee uint64) (*txs.Receipt, error) {

	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n", len(privKey))
	}
	trans.txMtx.Lock()
	defer trans.txMtx.Unlock()
	pa, err := acm.GeneratePrivateAccountFromPrivateKeyBytes(privKey)
	if err != nil {
		return nil, err
	}
	cache := trans.state // XXX: DON'T MUTATE THIS CACHE (used internally for CheckTx)
	acc, err := cache.GetAccount(pa.Address())
	if err != nil {
		return nil, err
	}
	sequence := uint64(1)
	if acc == nil {
		sequence = acc.Sequence() + uint64(1)
	}
	tx := txs.NewNameTxWithSequence(pa.PublicKey(), name, data, amount, fee, sequence)
	// Got ourselves a tx.
	txS, errS := trans.SignTx(tx, []acm.PrivateAccount{pa})
	if errS != nil {
		return nil, errS
	}
	return trans.BroadcastTx(txS)
}

// Sign a transaction
func (trans *transactor) SignTx(tx txs.Tx, privAccounts []acm.PrivateAccount) (txs.Tx, error) {
	// more checks?

	for i, privAccount := range privAccounts {
		if privAccount == nil || privAccount.PrivateKey().Unwrap() == nil {
			return nil, fmt.Errorf("invalid (empty) privAccount @%v", i)
		}
	}
	chainID := trans.blockchain.ChainID()
	switch tx.(type) {
	case *txs.NameTx:
		nameTx := tx.(*txs.NameTx)
		nameTx.Input.PublicKey = privAccounts[0].PublicKey()
		nameTx.Input.Signature = acm.ChainSign(privAccounts[0], chainID, nameTx)
	case *txs.SendTx:
		sendTx := tx.(*txs.SendTx)
		for i, input := range sendTx.Inputs {
			input.PublicKey = privAccounts[i].PublicKey()
			input.Signature = acm.ChainSign(privAccounts[i], chainID, sendTx)
		}
	case *txs.CallTx:
		callTx := tx.(*txs.CallTx)
		callTx.Input.PublicKey = privAccounts[0].PublicKey()
		callTx.Input.Signature = acm.ChainSign(privAccounts[0], chainID, callTx)
	case *txs.BondTx:
		bondTx := tx.(*txs.BondTx)
		// the first privaccount corresponds to the BondTx pub key.
		// the rest to the inputs
		bondTx.Signature = acm.ChainSign(privAccounts[0], chainID, bondTx)
		for i, input := range bondTx.Inputs {
			input.PublicKey = privAccounts[i+1].PublicKey()
			input.Signature = acm.ChainSign(privAccounts[i+1], chainID, bondTx)
		}
	case *txs.UnbondTx:
		unbondTx := tx.(*txs.UnbondTx)
		unbondTx.Signature = acm.ChainSign(privAccounts[0], chainID, unbondTx)
	case *txs.RebondTx:
		rebondTx := tx.(*txs.RebondTx)
		rebondTx.Signature = acm.ChainSign(privAccounts[0], chainID, rebondTx)
	default:
		return nil, fmt.Errorf("Object is not a proper transaction: %v\n", tx)
	}
	return tx, nil
}

func vmParams(blockchain blockchain.Blockchain) evm.Params {
	tip := blockchain.Tip()
	return evm.Params{
		BlockHeight: tip.LastBlockHeight(),
		BlockHash:   binary.LeftPadWord256(tip.LastBlockHash()),
		BlockTime:   tip.LastBlockTime().Unix(),
		GasLimit:    GasLimit,
	}
}
