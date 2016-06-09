package pipe

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/eris-ltd/eris-db/account"
	"github.com/eris-ltd/eris-db/manager/eris-mint/evm"
	"github.com/eris-ltd/eris-db/manager/eris-mint/state"
	"github.com/eris-ltd/eris-db/txs"

	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
	tEvents "github.com/tendermint/go-events"

	"github.com/eris-ltd/eris-db/tmsp"
)

type transactor struct {
	chainID      string
	eventSwitch  tEvents.Fireable
	erisdbApp    *tmsp.ErisDBApp
	eventEmitter EventEmitter
	txMtx        *sync.Mutex
}

func newTransactor(chainID string, eventSwitch tEvents.Fireable, erisdbApp *tmsp.ErisDBApp, eventEmitter EventEmitter) *transactor {
	txs := &transactor{
		chainID,
		eventSwitch,
		erisdbApp,
		eventEmitter,
		&sync.Mutex{},
	}
	return txs
}

// Run a contract's code on an isolated and unpersisted state
// Cannot be used to create new contracts
func (this *transactor) Call(fromAddress, toAddress, data []byte) (*Call, error) {

	cache := this.erisdbApp.GetCheckCache() // XXX: DON'T MUTATE THIS CACHE (used internally for CheckTx)
	outAcc := cache.GetAccount(toAddress)
	if outAcc == nil {
		return nil, fmt.Errorf("Account %X does not exist", toAddress)
	}
	if fromAddress == nil {
		fromAddress = []byte{}
	}
	callee := toVMAccount(outAcc)
	caller := &vm.Account{Address: cmn.LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	st := this.erisdbApp.GetState() // for block height, time
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   cmn.LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    10000000,
	}

	vmach := vm.NewVM(txCache, params, caller.Address, nil)
	vmach.SetFireable(this.eventSwitch)
	gas := int64(1000000000)
	ret, err := vmach.Call(caller, callee, callee.Code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	return &Call{Return: hex.EncodeToString(ret)}, nil
}

// Run the given code on an isolated and unpersisted state
// Cannot be used to create new contracts.
func (this *transactor) CallCode(fromAddress, code, data []byte) (*Call, error) {
	if fromAddress == nil {
		fromAddress = []byte{}
	}
	cache := this.erisdbApp.GetCheckCache() // XXX: DON'T MUTATE THIS CACHE (used internally for CheckTx)
	callee := &vm.Account{Address: cmn.LeftPadWord256(fromAddress)}
	caller := &vm.Account{Address: cmn.LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	st := this.erisdbApp.GetState() // for block height, time
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   cmn.LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    10000000,
	}

	vmach := vm.NewVM(txCache, params, caller.Address, nil)
	gas := int64(1000000000)
	ret, err := vmach.Call(caller, callee, code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	return &Call{Return: hex.EncodeToString(ret)}, nil
}

// Broadcast a transaction.
func (this *transactor) BroadcastTx(tx txs.Tx) (*Receipt, error) {

	err := this.erisdbApp.BroadcastTx(tx)
	if err != nil {
		return nil, fmt.Errorf("Error broadcasting transaction: %v", err)
	}

	txHash := txs.TxID(this.chainID, tx)
	var createsContract uint8
	var contractAddr []byte
	// check if creates new contract
	if callTx, ok := tx.(*txs.CallTx); ok {
		if len(callTx.Address) == 0 {
			createsContract = 1
			contractAddr = state.NewContractAddress(callTx.Input.Address, callTx.Input.Sequence)
		}
	}
	return &Receipt{txHash, createsContract, contractAddr}, nil
}

// Get all unconfirmed txs.
func (this *transactor) UnconfirmedTxs() (*UnconfirmedTxs, error) {
	// TODO-RPC
	return &UnconfirmedTxs{}, nil
}

// Orders calls to BroadcastTx using lock (waits for response from core before releasing)
func (this *transactor) Transact(privKey, address, data []byte, gasLimit, fee int64) (*Receipt, error) {
	var addr []byte
	if len(address) == 0 {
		addr = nil
	} else if len(address) != 20 {
		return nil, fmt.Errorf("Address is not of the right length: %d\n", len(address))
	} else {
		addr = address
	}
	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n", len(privKey))
	}
	this.txMtx.Lock()
	defer this.txMtx.Unlock()
	pa := account.GenPrivAccountFromPrivKeyBytes(privKey)
	cache := this.erisdbApp.GetCheckCache() // XXX: DON'T MUTATE THIS CACHE (used internally for CheckTx)
	acc := cache.GetAccount(pa.Address)
	var sequence int
	if acc == nil {
		sequence = 1
	} else {
		sequence = acc.Sequence + 1
	}
	// fmt.Printf("Sequence %d\n", sequence)
	txInput := &txs.TxInput{
		Address:  pa.Address,
		Amount:   1,
		Sequence: sequence,
		PubKey:   pa.PubKey,
	}
	tx := &txs.CallTx{
		Input:    txInput,
		Address:  addr,
		GasLimit: gasLimit,
		Fee:      fee,
		Data:     data,
	}

	// Got ourselves a tx.
	txS, errS := this.SignTx(tx, []*account.PrivAccount{pa})
	if errS != nil {
		return nil, errS
	}
	return this.BroadcastTx(txS)
}

func (this *transactor) TransactAndHold(privKey, address, data []byte, gasLimit, fee int64) (*txs.EventDataCall, error) {
	rec, tErr := this.Transact(privKey, address, data, gasLimit, fee)
	if tErr != nil {
		return nil, tErr
	}
	var addr []byte
	if rec.CreatesContract == 1 {
		addr = rec.ContractAddr
	} else {
		addr = address
	}
	wc := make(chan *txs.EventDataCall)
	subId := fmt.Sprintf("%X", rec.TxHash)
	this.eventEmitter.Subscribe(subId, txs.EventStringAccCall(addr), func(evt tEvents.EventData) {
		event := evt.(txs.EventDataCall)
		if bytes.Equal(event.TxID, rec.TxHash) {
			wc <- &event
		}
	})

	timer := time.NewTimer(300 * time.Second)
	toChan := timer.C

	var ret *txs.EventDataCall
	var rErr error

	select {
	case <-toChan:
		rErr = fmt.Errorf("Transaction timed out. Hash: " + subId)
	case e := <-wc:
		timer.Stop()
		if e.Exception != "" {
			rErr = fmt.Errorf("Error when transacting: " + e.Exception)
		} else {
			ret = e
		}
	}
	this.eventEmitter.Unsubscribe(subId)
	return ret, rErr
}

func (this *transactor) TransactNameReg(privKey []byte, name, data string, amount, fee int64) (*Receipt, error) {

	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n", len(privKey))
	}
	this.txMtx.Lock()
	defer this.txMtx.Unlock()
	pa := account.GenPrivAccountFromPrivKeyBytes(privKey)
	cache := this.erisdbApp.GetCheckCache() // XXX: DON'T MUTATE THIS CACHE (used internally for CheckTx)
	acc := cache.GetAccount(pa.Address)
	var sequence int
	if acc == nil {
		sequence = 1
	} else {
		sequence = acc.Sequence + 1
	}
	tx := txs.NewNameTxWithNonce(pa.PubKey, name, data, amount, fee, sequence)
	// Got ourselves a tx.
	txS, errS := this.SignTx(tx, []*account.PrivAccount{pa})
	if errS != nil {
		return nil, errS
	}
	return this.BroadcastTx(txS)
}

// Sign a transaction
func (this *transactor) SignTx(tx txs.Tx, privAccounts []*account.PrivAccount) (txs.Tx, error) {
	// more checks?

	for i, privAccount := range privAccounts {
		if privAccount == nil || privAccount.PrivKey == nil {
			return nil, fmt.Errorf("Invalid (empty) privAccount @%v", i)
		}
	}
	switch tx.(type) {
	case *txs.NameTx:
		nameTx := tx.(*txs.NameTx)
		nameTx.Input.PubKey = privAccounts[0].PubKey
		nameTx.Input.Signature = privAccounts[0].Sign(this.chainID, nameTx)
	case *txs.SendTx:
		sendTx := tx.(*txs.SendTx)
		for i, input := range sendTx.Inputs {
			input.PubKey = privAccounts[i].PubKey
			input.Signature = privAccounts[i].Sign(this.chainID, sendTx)
		}
		break
	case *txs.CallTx:
		callTx := tx.(*txs.CallTx)
		callTx.Input.PubKey = privAccounts[0].PubKey
		callTx.Input.Signature = privAccounts[0].Sign(this.chainID, callTx)
		break
	case *txs.BondTx:
		bondTx := tx.(*txs.BondTx)
		// the first privaccount corresponds to the BondTx pub key.
		// the rest to the inputs
		bondTx.Signature = privAccounts[0].Sign(this.chainID, bondTx).(crypto.SignatureEd25519)
		for i, input := range bondTx.Inputs {
			input.PubKey = privAccounts[i+1].PubKey
			input.Signature = privAccounts[i+1].Sign(this.chainID, bondTx)
		}
		break
	case *txs.UnbondTx:
		unbondTx := tx.(*txs.UnbondTx)
		unbondTx.Signature = privAccounts[0].Sign(this.chainID, unbondTx).(crypto.SignatureEd25519)
		break
	case *txs.RebondTx:
		rebondTx := tx.(*txs.RebondTx)
		rebondTx.Signature = privAccounts[0].Sign(this.chainID, rebondTx).(crypto.SignatureEd25519)
		break
	default:
		return nil, fmt.Errorf("Object is not a proper transaction: %v\n", tx)
	}
	return tx, nil
}

// No idea what this does.
func toVMAccount(acc *account.Account) *vm.Account {
	return &vm.Account{
		Address: cmn.LeftPadWord256(acc.Address),
		Balance: acc.Balance,
		Code:    acc.Code,
		Nonce:   int64(acc.Sequence),
		Other:   acc.PubKey,
	}
}
