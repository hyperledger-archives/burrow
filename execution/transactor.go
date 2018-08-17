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

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/consensus/tendermint/codes"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

const (
	BlockingTimeout     = 10 * time.Second
	SubscribeBufferSize = 10
)

// Transactor is responsible for helping to formulate, sign, and broadcast transactions to tendermint
//
// The BroadcastTx* methods are able to work against the mempool Accounts (pending) state rather than the
// committed (final) Accounts state and can assign a sequence number based on all of the txs
// seen since the last block - provided these transactions are successfully committed (via DeliverTx) then
// subsequent transactions will have valid sequence numbers. This allows Burrow to coordinate sequencing and signing
// for a key it holds or is provided - it is down to the key-holder to manage the mutual information between transactions
// concurrent within a new block window.
type Transactor struct {
	Tip             bcm.BlockchainInfo
	Subscribable    event.Subscribable
	MempoolAccounts *Accounts
	checkTxAsync    func(tx tmTypes.Tx, cb func(*abciTypes.Response)) error
	txEncoder       txs.Encoder
	logger          *logging.Logger
}

func NewTransactor(tip bcm.BlockchainInfo, subscribable event.Subscribable, mempoolAccounts *Accounts,
	checkTxAsync func(tx tmTypes.Tx, cb func(*abciTypes.Response)) error, txEncoder txs.Encoder,
	logger *logging.Logger) *Transactor {

	return &Transactor{
		Tip:             tip,
		Subscribable:    subscribable,
		MempoolAccounts: mempoolAccounts,
		checkTxAsync:    checkTxAsync,
		txEncoder:       txEncoder,
		logger:          logger.With(structure.ComponentKey, "Transactor"),
	}
}

func (trans *Transactor) BroadcastTxSync(ctx context.Context, txEnv *txs.Envelope) (*exec.TxExecution, error) {
	// Sign unless already signed - note we must attempt signing before subscribing so we get accurate final TxHash
	unlock, err := trans.MaybeSignTxMempool(txEnv)
	if err != nil {
		return nil, err
	}
	// We will try and call this before the function exits unless we error but it is idempotent
	defer unlock()
	// Subscribe before submitting to mempool
	txHash := txEnv.Tx.Hash()
	subID := event.GenSubID()
	out, err := trans.Subscribable.Subscribe(ctx, subID, exec.QueryForTxExecution(txHash), SubscribeBufferSize)
	if err != nil {
		// We do not want to hold the lock with a defer so we must
		return nil, err
	}
	// Push Tx to mempool
	checkTxReceipt, err := trans.CheckTxSync(txEnv)
	unlock()
	if err != nil {
		return nil, err
	}
	defer trans.Subscribable.UnsubscribeAll(context.Background(), subID)
	// Wait for all responses
	timer := time.NewTimer(BlockingTimeout)
	defer timer.Stop()

	// Get all the execution events for this Tx
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, fmt.Errorf("timed out waiting for transaction with hash %v timed out after %v",
			checkTxReceipt.TxHash, BlockingTimeout)
	case msg := <-out:
		txe := msg.(*exec.TxExecution)
		if txe.Exception != nil && txe.Exception.ErrorCode() != errors.ErrorCodeExecutionReverted {
			return nil, errors.Wrap(txe.Exception, "exception during transaction execution")
		}
		return txe, nil
	}
}

// Broadcast a transaction without waiting for confirmation - will attempt to sign server-side and set sequence numbers
// if no signatures are provided
func (trans *Transactor) BroadcastTxAsync(txEnv *txs.Envelope) (*txs.Receipt, error) {
	return trans.CheckTxSync(txEnv)
}

// Broadcast a transaction and waits for a response from the mempool. Transactions to BroadcastTx will block during
// various mempool operations (managed by Tendermint) including mempool Reap, Commit, and recheckTx.
func (trans *Transactor) CheckTxSync(txEnv *txs.Envelope) (*txs.Receipt, error) {
	trans.logger.Trace.Log("method", "CheckTxSync",
		"tx_hash", txEnv.Tx.Hash(),
		"tx", txEnv.String())
	// Sign unless already signed
	unlock, err := trans.MaybeSignTxMempool(txEnv)
	if err != nil {
		return nil, err
	}
	defer unlock()
	err = txEnv.Validate()
	if err != nil {
		return nil, err
	}
	txBytes, err := trans.txEncoder.EncodeTx(txEnv)
	if err != nil {
		return nil, err
	}
	return trans.CheckTxSyncRaw(txBytes)
}

func (trans *Transactor) MaybeSignTxMempool(txEnv *txs.Envelope) (UnlockFunc, error) {
	// Sign unless already signed
	if len(txEnv.Signatories) == 0 {
		var err error
		var unlock UnlockFunc
		// We are writing signatures back to txEnv so don't shadow txEnv here
		txEnv, unlock, err = trans.SignTxMempool(txEnv)
		if err != nil {
			return nil, fmt.Errorf("error signing transaction: %v", err)
		}
		// Hash will have change since we signed
		txEnv.Tx.Rehash()

		// Make this idempotent for defer
		var once sync.Once
		return func() { once.Do(unlock) }, nil
	}
	return func() {}, nil
}

func (trans *Transactor) SignTxMempool(txEnv *txs.Envelope) (*txs.Envelope, UnlockFunc, error) {
	inputs := txEnv.Tx.GetInputs()
	signers := make([]acm.AddressableSigner, len(inputs))
	unlockers := make([]UnlockFunc, len(inputs))
	for i, input := range inputs {
		ssa, err := trans.MempoolAccounts.SequentialSigningAccount(input.Address)
		if err != nil {
			return nil, nil, err
		}
		sa, unlock, err := ssa.Lock()
		if err != nil {
			return nil, nil, err
		}
		// Hold lock until safely in mempool - important that this is held until after CheckTxSync returns
		unlockers[i] = unlock
		signers[i] = sa
		// Set sequence number consecutively from mempool
		input.Sequence = sa.Sequence() + 1
	}

	err := txEnv.Sign(signers...)
	if err != nil {
		return nil, nil, err
	}
	return txEnv, UnlockFunc(func() {
		for _, unlock := range unlockers {
			defer unlock()
		}
	}), nil
}

func (trans *Transactor) SignTx(txEnv *txs.Envelope) (*txs.Envelope, error) {
	var err error
	inputs := txEnv.Tx.GetInputs()
	signers := make([]acm.AddressableSigner, len(inputs))
	for i, input := range inputs {
		signers[i], err = trans.MempoolAccounts.SigningAccount(input.Address)
	}
	err = txEnv.Sign(signers...)
	if err != nil {
		return nil, err
	}
	return txEnv, nil
}

func (trans *Transactor) CheckTxSyncRaw(txBytes []byte) (*txs.Receipt, error) {
	responseCh := make(chan *abciTypes.Response, 1)
	err := trans.CheckTxAsyncRaw(txBytes, func(res *abciTypes.Response) {
		responseCh <- res
	})
	if err != nil {
		return nil, err
	}
	timer := time.NewTimer(BlockingTimeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil, fmt.Errorf("timed out waiting for CheckTx response in CheckTxSyncRaw")
	case response := <-responseCh:
		checkTxResponse := response.GetCheckTx()
		if checkTxResponse == nil {
			return nil, fmt.Errorf("application did not return CheckTx response")
		}

		switch checkTxResponse.Code {
		case codes.TxExecutionSuccessCode:
			receipt, err := txs.DecodeReceipt(checkTxResponse.Data)
			if err != nil {
				return nil, fmt.Errorf("could not deserialise transaction receipt: %s", err)
			}
			return receipt, nil
		default:
			return nil, errors.ErrorCodef(errors.Code(checkTxResponse.Code),
				"error returned by Tendermint in BroadcastTxSync ABCI log: %v", checkTxResponse.Log)
		}
	}
}

func (trans *Transactor) CheckTxAsyncRaw(txBytes []byte, callback func(res *abciTypes.Response)) error {
	return trans.checkTxAsync(txBytes, callback)
}

func (trans *Transactor) CheckTxAsync(txEnv *txs.Envelope, callback func(res *abciTypes.Response)) error {
	err := txEnv.Validate()
	if err != nil {
		return err
	}
	txBytes, err := trans.txEncoder.EncodeTx(txEnv)
	if err != nil {
		return fmt.Errorf("error encoding transaction: %v", err)
	}
	return trans.CheckTxAsyncRaw(txBytes, callback)
}

func (trans *Transactor) CallCodeSim(fromAddress crypto.Address, code, data []byte) (*exec.TxExecution, error) {
	return CallCodeSim(trans.MempoolAccounts, trans.Tip, fromAddress, fromAddress, code, data, trans.logger)
}

func (trans *Transactor) CallSim(fromAddress, address crypto.Address, data []byte) (*exec.TxExecution, error) {
	return CallSim(trans.MempoolAccounts, trans.Tip, fromAddress, address, data, trans.logger)
}
