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
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/executors"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	abciTypes "github.com/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

const BlockingTimeoutSeconds = 30

type Call struct {
	Return  binary.HexBytes
	GasUsed uint64
}

// Transactor is the controller/middleware for the v0 RPC
type Transactor struct {
	tip              *blockchain.Tip
	eventEmitter     event.Emitter
	broadcastTxAsync func(tx tmTypes.Tx, cb func(*abciTypes.Response)) error
	txEncoder        txs.Encoder
	logger           *logging.Logger
}

func NewTransactor(tip *blockchain.Tip, eventEmitter event.Emitter,
	broadcastTxAsync func(tx tmTypes.Tx, cb func(*abciTypes.Response)) error, txEncoder txs.Encoder,
	logger *logging.Logger) *Transactor {

	return &Transactor{
		tip:              tip,
		eventEmitter:     eventEmitter,
		broadcastTxAsync: broadcastTxAsync,
		txEncoder:        txEncoder,
		logger:           logger.With(structure.ComponentKey, "Transactor"),
	}
}

// Run a contract's code on an isolated and unpersisted state
// Cannot be used to create new contracts
func (trans *Transactor) Call(reader state.Reader, fromAddress, address crypto.Address,
	data []byte) (call *Call, err error) {

	if evm.IsRegisteredNativeContract(address.Word256()) {
		return nil, fmt.Errorf("attempt to call native contract at address "+
			"%X, but native contracts can not be called directly. Use a deployed "+
			"contract that calls the native function instead", address)
	}
	// This was being run against CheckTx cache, need to understand the reasoning
	callee, err := state.GetMutableAccount(reader, address)
	if err != nil {
		return nil, err
	}
	if callee == nil {
		return nil, fmt.Errorf("account %s does not exist", address)
	}
	caller := acm.ConcreteAccount{Address: fromAddress}.MutableAccount()
	txCache := state.NewCache(reader)
	params := vmParams(trans.tip)

	vmach := evm.NewVM(params, caller.Address(), nil, trans.logger.WithScope("Call"))
	vmach.SetPublisher(trans.eventEmitter)

	gas := params.GasLimit
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic from VM in simulated call: %v\n%s", r, debug.Stack())
		}
	}()
	ret, err := vmach.Call(txCache, caller, callee, callee.Code(), data, 0, &gas)
	if err != nil {
		return nil, err
	}
	gasUsed := params.GasLimit - gas
	return &Call{Return: ret, GasUsed: gasUsed}, nil
}

// Run the given code on an isolated and unpersisted state
// Cannot be used to create new contracts.
func (trans *Transactor) CallCode(reader state.Reader, fromAddress crypto.Address, code, data []byte) (*Call, error) {
	// This was being run against CheckTx cache, need to understand the reasoning
	callee := acm.ConcreteAccount{Address: fromAddress}.MutableAccount()
	caller := acm.ConcreteAccount{Address: fromAddress}.MutableAccount()
	txCache := state.NewCache(reader)
	params := vmParams(trans.tip)

	vmach := evm.NewVM(params, caller.Address(), nil, trans.logger.WithScope("CallCode"))
	gas := params.GasLimit
	ret, err := vmach.Call(txCache, caller, callee, code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	gasUsed := params.GasLimit - gas
	return &Call{Return: ret, GasUsed: gasUsed}, nil
}

func (trans *Transactor) BroadcastTxAsyncRaw(txBytes []byte, callback func(res *abciTypes.Response)) error {
	return trans.broadcastTxAsync(txBytes, callback)
}

func (trans *Transactor) BroadcastTxAsync(txEnv *txs.Envelope, callback func(res *abciTypes.Response)) error {
	err := txEnv.Validate()
	if err != nil {
		return err
	}
	txBytes, err := trans.txEncoder.EncodeTx(txEnv)
	if err != nil {
		return fmt.Errorf("error encoding transaction: %v", err)
	}
	return trans.BroadcastTxAsyncRaw(txBytes, callback)
}

// Broadcast a transaction and waits for a response from the mempool. Transactions to BroadcastTx will block during
// various mempool operations (managed by Tendermint) including mempool Reap, Commit, and recheckTx.
func (trans *Transactor) BroadcastTx(txEnv *txs.Envelope) (*txs.Receipt, error) {
	trans.logger.Trace.Log("method", "BroadcastTx",
		"tx_hash", txEnv.Tx.Hash(),
		"tx", txEnv.String())
	err := txEnv.Validate()
	if err != nil {
		return nil, err
	}
	txBytes, err := trans.txEncoder.EncodeTx(txEnv)
	if err != nil {
		return nil, err
	}
	return trans.BroadcastTxRaw(txBytes)
}

func (trans *Transactor) BroadcastTxRaw(txBytes []byte) (*txs.Receipt, error) {
	responseCh := make(chan *abciTypes.Response, 1)
	err := trans.BroadcastTxAsyncRaw(txBytes, func(res *abciTypes.Response) {
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
		err := json.Unmarshal(checkTxResponse.Data, receipt)
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
func (trans *Transactor) Transact(sequentialSigningAccount *SequentialSigningAccount, address *crypto.Address, data []byte,
	gasLimit, value, fee uint64) (*txs.Receipt, error) {

	// Use the get the freshest sequence numbers from mempool state and hold the lock until we get a response from
	// CheckTx
	inputAccount, unlock, err := sequentialSigningAccount.Lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	if binary.IsUint64SumOverflow(value, fee) {
		return nil, fmt.Errorf("amount to transfer overflows uint64 amount = %v (value) + %v (fee)", value, fee)
	}
	txEnv, err := trans.formulateCallTx(inputAccount, address, data, gasLimit, value+fee, fee)
	if err != nil {
		return nil, err
	}
	return trans.BroadcastTx(txEnv)
}

func (trans *Transactor) TransactAndHold(ctx context.Context, sequentialSigningAccount *SequentialSigningAccount,
	address *crypto.Address, data []byte, gasLimit, value, fee uint64) (*events.EventDataCall, error) {

	inputAccount, unlock, err := sequentialSigningAccount.Lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	if binary.IsUint64SumOverflow(value, fee) {
		return nil, fmt.Errorf("amount to transfer overflows uint64 amount = %v (value) + %v (fee)", value, fee)
	}
	callTxEnv, err := trans.formulateCallTx(inputAccount, address, data, gasLimit, value+fee, fee)
	if err != nil {
		return nil, err
	}

	expectedReceipt := callTxEnv.Tx.GenerateReceipt()

	subID, err := event.GenerateSubscriptionID()
	if err != nil {
		return nil, err
	}

	// We want non-blocking on the first event received (but buffer the value),
	// after which we want to block (and then discard the value - see below)
	ch := make(chan *events.EventDataCall, 1)

	err = events.SubscribeAccountCall(context.Background(), trans.eventEmitter, subID, expectedReceipt.ContractAddress,
		expectedReceipt.TxHash, 0, ch)
	if err != nil {
		return nil, err
	}
	// Will clean up callback goroutine and subscription in pubsub
	defer trans.eventEmitter.UnsubscribeAll(context.Background(), subID)

	receipt, err := trans.BroadcastTx(callTxEnv)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(receipt.TxHash, expectedReceipt.TxHash) {
		return nil, fmt.Errorf("BroadcastTx received TxHash %X but %X was expected",
			receipt.TxHash, expectedReceipt.TxHash)
	}

	timer := time.NewTimer(BlockingTimeoutSeconds * time.Second)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, fmt.Errorf("transaction timed out TxHash: %X", expectedReceipt.TxHash)
	case eventDataCall := <-ch:
		if eventDataCall.Exception != nil && eventDataCall.Exception.ErrorCode() != errors.ErrorCodeExecutionReverted {
			return nil, fmt.Errorf("error when transacting: %v", eventDataCall.Exception)
		} else {
			return eventDataCall, nil
		}
	}
}
func (trans *Transactor) formulateCallTx(inputAccount *SigningAccount, address *crypto.Address, data []byte,
	gasLimit, amount, fee uint64) (*txs.Envelope, error) {

	txInput := &payload.TxInput{
		Address:  inputAccount.Address(),
		Amount:   amount,
		Sequence: inputAccount.Sequence() + 1,
	}
	tx := &payload.CallTx{
		Input:    txInput,
		Address:  address,
		GasLimit: gasLimit,
		Fee:      fee,
		Data:     data,
	}

	env := txs.Enclose(trans.tip.ChainID(), tx)
	// Got ourselves a tx.
	err := env.Sign(inputAccount)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func (trans *Transactor) Send(sequentialSigningAccount *SequentialSigningAccount, toAddress crypto.Address,
	amount uint64) (*txs.Receipt, error) {

	inputAccount, unlock, err := sequentialSigningAccount.Lock()
	if err != nil {
		return nil, err
	}
	defer unlock()

	sendTxEnv, err := trans.formulateSendTx(inputAccount, toAddress, amount)
	if err != nil {
		return nil, err
	}

	return trans.BroadcastTx(sendTxEnv)
}

func (trans *Transactor) SendAndHold(ctx context.Context, sequentialSigningAccount *SequentialSigningAccount,
	toAddress crypto.Address, amount uint64) (*txs.Receipt, error) {

	inputAccount, unlock, err := sequentialSigningAccount.Lock()
	if err != nil {
		return nil, err
	}
	defer unlock()

	sendTxEnv, err := trans.formulateSendTx(inputAccount, toAddress, amount)
	if err != nil {
		return nil, err
	}
	expectedReceipt := sendTxEnv.Tx.GenerateReceipt()

	subID, err := event.GenerateSubscriptionID()
	if err != nil {
		return nil, err
	}

	wc := make(chan *payload.SendTx)
	err = events.SubscribeAccountOutputSendTx(context.Background(), trans.eventEmitter, subID, toAddress,
		expectedReceipt.TxHash, wc)
	if err != nil {
		return nil, err
	}
	defer trans.eventEmitter.UnsubscribeAll(context.Background(), subID)

	receipt, err := trans.BroadcastTx(sendTxEnv)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(receipt.TxHash, expectedReceipt.TxHash) {
		return nil, fmt.Errorf("BroadcastTx received TxHash %X but %X was expected",
			receipt.TxHash, expectedReceipt.TxHash)
	}

	timer := time.NewTimer(BlockingTimeoutSeconds * time.Second)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, fmt.Errorf("transaction timed out TxHash: %X", expectedReceipt.TxHash)
	case sendTx := <-wc:
		// This is a double check - we subscribed to this tx's hash so something has gone wrong if the amounts don't match
		if sendTx.Inputs[0].Amount == amount {
			return expectedReceipt, nil
		}
		return nil, fmt.Errorf("received SendTx but hash doesn't seem to match what we subscribed to, "+
			"received SendTx: %v which does not match receipt on sending: %v", sendTx, expectedReceipt)
	}
}

func (trans *Transactor) formulateSendTx(inputAccount *SigningAccount, toAddress crypto.Address,
	amount uint64) (*txs.Envelope, error) {

	sendTx := payload.NewSendTx()
	txInput := &payload.TxInput{
		Address:  inputAccount.Address(),
		Amount:   amount,
		Sequence: inputAccount.Sequence() + 1,
	}
	sendTx.Inputs = append(sendTx.Inputs, txInput)
	txOutput := &payload.TxOutput{Address: toAddress, Amount: amount}
	sendTx.Outputs = append(sendTx.Outputs, txOutput)

	env := txs.Enclose(trans.tip.ChainID(), sendTx)
	err := env.Sign(inputAccount)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (trans *Transactor) TransactNameReg(sequentialSigningAccount *SequentialSigningAccount, name, data string, amount,
	fee uint64) (*txs.Receipt, error) {

	inputAccount, unlock, err := sequentialSigningAccount.Lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	// Formulate and sign
	tx := payload.NewNameTxWithSequence(inputAccount.PublicKey(), name, data, amount, fee, inputAccount.Sequence()+1)
	env := txs.Enclose(trans.tip.ChainID(), tx)
	err = env.Sign(inputAccount)
	if err != nil {
		return nil, err
	}
	return trans.BroadcastTx(env)
}

// Sign a transaction
func (trans *Transactor) SignTx(txEnv *txs.Envelope, signingAccounts []acm.AddressableSigner) (*txs.Envelope, error) {
	// more checks?
	err := txEnv.Sign(signingAccounts...)
	if err != nil {
		return nil, err
	}
	return txEnv, nil
}

func vmParams(tip *blockchain.Tip) evm.Params {
	return evm.Params{
		BlockHeight: tip.LastBlockHeight(),
		BlockHash:   binary.LeftPadWord256(tip.LastBlockHash()),
		BlockTime:   tip.LastBlockTime().Unix(),
		GasLimit:    executors.GasLimit,
	}
}
