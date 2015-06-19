package pipe

import (
	"encoding/hex"
	"fmt"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	cmn "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	cs "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/consensus"
	mempl "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/mempool"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/vm"
	"sync"
)

const (
	DEFAULT_BLOCKS_WAIT = 10
	SUB_ID              = "TransactorSubBlock"
	EVENT_ID            = "NewBlock"
)

type transactor struct {
	consensusState *cs.ConsensusState
	mempoolReactor *mempl.MempoolReactor
	pending        []TxFuture
	pendingLock    *sync.Mutex
	eventEmitter   EventEmitter
}

func newTransactor(consensusState *cs.ConsensusState, mempoolReactor *mempl.MempoolReactor, eventEmitter EventEmitter) *transactor {
	txs := &transactor{
		consensusState,
		mempoolReactor,
		[]TxFuture{},
		&sync.Mutex{},
		eventEmitter,
	}
	/*
		eventEmitter.Subscribe(SUB_ID, EVENT_ID, func(v interface{}) {
			block := v.(*types.Block)
			for _, fut := range txs.pending {
				fut.NewBlock(block)
			}
		})
	*/
	return txs
}

// Run a contract's code on an isolated and unpersisted state
// Cannot be used to create new contracts
func (this *transactor) Call(address, data []byte) (*Call, error) {

	st := this.consensusState.GetState() // performs a copy
	cache := state.NewBlockCache(st)
	outAcc := cache.GetAccount(address)
	if outAcc == nil {
		return nil, fmt.Errorf("Account %x does not exist", address)
	}
	callee := toVMAccount(outAcc)
	caller := &vm.Account{Address: cmn.Zero256}
	txCache := state.NewTxCache(cache)
	params := vm.Params{
		BlockHeight: uint64(st.LastBlockHeight),
		BlockHash:   cmn.LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    10000000,
	}

	vmach := vm.NewVM(txCache, params, caller.Address, nil)
	gas := uint64(1000000000)
	ret, err := vmach.Call(caller, callee, callee.Code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	return &Call{Return: hex.EncodeToString(ret)}, nil
}

// Run the given code on an isolated and unpersisted state
// Cannot be used to create new contracts.
func (this *transactor) CallCode(code, data []byte) (*Call, error) {

	st := this.consensusState.GetState() // performs a copy
	cache := this.mempoolReactor.Mempool.GetCache()
	callee := &vm.Account{Address: cmn.Zero256}
	caller := &vm.Account{Address: cmn.Zero256}
	txCache := state.NewTxCache(cache)
	params := vm.Params{
		BlockHeight: uint64(st.LastBlockHeight),
		BlockHash:   cmn.LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    10000000,
	}

	vmach := vm.NewVM(txCache, params, caller.Address, nil)
	gas := uint64(1000000000)
	ret, err := vmach.Call(caller, callee, code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	return &Call{Return: hex.EncodeToString(ret)}, nil
}

// Broadcast a transaction.
func (this *transactor) BroadcastTx(tx types.Tx) (*Receipt, error) {
	err := this.mempoolReactor.BroadcastTx(tx)
	if err != nil {
		return nil, fmt.Errorf("Error broadcasting transaction: %v", err)
	}
	chainId := config.GetString("chain_id")
	txHash := types.TxId(chainId, tx)
	var createsContract uint8
	var contractAddr []byte
	// check if creates new contract
	if callTx, ok := tx.(*types.CallTx); ok {
		if len(callTx.Address) == 0 {
			createsContract = 1
			contractAddr = state.NewContractAddress(callTx.Input.Address, uint64(callTx.Input.Sequence))
		}
	}
	return &Receipt{txHash, createsContract, contractAddr}, nil
}

// Get all unconfirmed txs.
func (this *transactor) UnconfirmedTxs() (*UnconfirmedTxs, error) {
	transactions := this.mempoolReactor.Mempool.GetProposalTxs()
	return &UnconfirmedTxs{transactions}, nil
}

func (this *transactor) TransactAsync(privKey, address, data []byte, gasLimit, fee uint64) (*TransactionResult, error) {
	return nil, nil
}

func (this *transactor) Transact(privKey, address, data []byte, gasLimit, fee uint64) (*Receipt, error) {
	fmt.Printf("ADDRESS: %v\n", address)
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

	key := [64]byte{}
	copy(key[:], privKey[0:64])
	pa := account.GenPrivAccountFromKey(key)
	cache := this.mempoolReactor.Mempool.GetCache()
	acc := cache.GetAccount(pa.Address)
	var sequence uint
	if acc == nil {
		sequence = 1
	} else {
		sequence = acc.Sequence + 1
	}
	fmt.Printf("NONCE: %d\n", sequence)
	txInput := &types.TxInput{
		Address:  pa.Address,
		Amount:   1000,
		Sequence: sequence,
		PubKey:   pa.PubKey,
	}
	tx := &types.CallTx{
		Input:    txInput,
		Address:  addr,
		GasLimit: 1000,
		Fee:      1000,
		Data:     data,
	}
	// Got ourselves a tx.
	txS, errS := this.SignTx(tx, []*account.PrivAccount{pa})
	if errS != nil {
		return nil, errS
	}
	return this.BroadcastTx(txS)
}

// Sign a transaction
func (this *transactor) SignTx(tx types.Tx, privAccounts []*account.PrivAccount) (types.Tx, error) {
	// more checks?

	for i, privAccount := range privAccounts {
		if privAccount == nil || privAccount.PrivKey == nil {
			return nil, fmt.Errorf("Invalid (empty) privAccount @%v", i)
		}
	}
	chainId := config.GetString("chain_id")
	switch tx.(type) {
	case *types.SendTx:
		sendTx := tx.(*types.SendTx)
		for i, input := range sendTx.Inputs {
			input.PubKey = privAccounts[i].PubKey
			input.Signature = privAccounts[i].Sign(chainId, sendTx)
		}
		break
	case *types.CallTx:
		callTx := tx.(*types.CallTx)
		callTx.Input.PubKey = privAccounts[0].PubKey
		callTx.Input.Signature = privAccounts[0].Sign(chainId, callTx)
		break
	case *types.BondTx:
		bondTx := tx.(*types.BondTx)
		// the first privaccount corresponds to the BondTx pub key.
		// the rest to the inputs
		bondTx.Signature = privAccounts[0].Sign(chainId, bondTx).(account.SignatureEd25519)
		for i, input := range bondTx.Inputs {
			input.PubKey = privAccounts[i+1].PubKey
			input.Signature = privAccounts[i+1].Sign(chainId, bondTx)
		}
		break
	case *types.UnbondTx:
		unbondTx := tx.(*types.UnbondTx)
		unbondTx.Signature = privAccounts[0].Sign(chainId, unbondTx).(account.SignatureEd25519)
		break
	case *types.RebondTx:
		rebondTx := tx.(*types.RebondTx)
		rebondTx.Signature = privAccounts[0].Sign(chainId, rebondTx).(account.SignatureEd25519)
		break
	default:
		return nil, fmt.Errorf("Object is not a proper transaction: %v\n", tx)
	}
	return tx, nil
}

// No idea what this does.
func toVMAccount(acc *account.Account) *vm.Account {
	return &vm.Account{
		Address:     cmn.LeftPadWord256(acc.Address),
		Balance:     acc.Balance,
		Code:        acc.Code, // This is crazy.
		Nonce:       uint64(acc.Sequence),
		StorageRoot: cmn.LeftPadWord256(acc.StorageRoot),
		Other:       acc.PubKey,
	}
}

// This is the different status codes for transactions.
// 0 - the tx tracker object is being set up.
// 1 - the tx has been created and passed into the tx pool.
// 2 - the tx was succesfully committed into a block.
// Errors
// -1 - the tx failed.
const (
	TX_NEW_CODE      int8 = 0
	TX_POOLED_CODE   int8 = 1
	TX_COMITTED_CODE int8 = 2
	TX_FAILED_CODE   int8 = -1
)

// Number of bytes in a transaction hash
const TX_HASH_BYTES = 32

// Length of the tx hash hex-string (prepended by 0x)
const TX_HASH_LENGTH = 2 * TX_HASH_BYTES

type TxFuture interface {
	// Tx Hash
	Hash() string
	// Target account.
	Target() string
	// Get the Receipt for this transaction.
	Results() *TransactionResult
	// This will block and wait for the tx to be done.
	Get() *TransactionResult
	// This will block for 'timeout' miliseconds and wait for
	// the tx to be done. 0 means no timeout, and is equivalent
	// to calling 'Get()'.
	GetWithTimeout(timeout uint64) *TransactionResult
	// Checks the status. The status codes can be find near the
	// top of this file.
	StatusCode() int8
	// This is true when the transaction is done (whether it was successful or not).
	Done() bool
}

// Implements the 'TxFuture' interface.
type TxFutureImpl struct {
	receipt    *Receipt
	result     *TransactionResult
	target     string
	status     int8
	transactor Transactor
	errStr     string
	getLock    *sync.Mutex
}

func (this *TxFutureImpl) Results() *TransactionResult {
	return this.result
}

func (this *TxFutureImpl) StatusCode() int8 {
	return this.status
}

func (this *TxFutureImpl) Done() bool {
	return this.status == TX_COMITTED_CODE || this.status == TX_FAILED_CODE
}

func (this *TxFutureImpl) Wait() *TransactionResult {
	return this.WaitWithTimeout(0)
}

// We wait for blocks, and when a block arrives we check if tx is committed.
// This will return after it has been confirmed that tx was committed, or if
// it failed, and for a maximum of 'blocks' blocks. If 'blocks' is set to 0,
// it will be set to DEFAULT_BLOCKS_WAIT.
// This is a temporary solution until we have solidity events.
func (this *TxFutureImpl) WaitWithTimeout(blocks int) *TransactionResult {
	return nil
}

func (this *TxFutureImpl) setStatus(status int8, errorStr string) {
	this.status = status
	if status == TX_FAILED_CODE {
		this.errStr = errorStr
	}
}
