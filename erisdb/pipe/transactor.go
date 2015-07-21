package pipe

import (
	"encoding/hex"
	"fmt"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	cmn "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	cs "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/consensus"
	tEvents "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/events"
	mempl "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/mempool"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/vm"
)

const (
	DEFAULT_BLOCKS_WAIT = 10
	SUB_ID              = "TransactorSubBlock"
	EVENT_ID            = "NewBlock"
)

type transactor struct {
	eventSwitch    tEvents.Fireable
	consensusState *cs.ConsensusState
	mempoolReactor *mempl.MempoolReactor
	eventEmitter   EventEmitter
}

func newTransactor(eventSwitch tEvents.Fireable, consensusState *cs.ConsensusState, mempoolReactor *mempl.MempoolReactor, eventEmitter EventEmitter) *transactor {
	txs := &transactor{
		eventSwitch,
		consensusState,
		mempoolReactor,
		eventEmitter,
	}
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
func (this *transactor) CallCode(code, data []byte) (*Call, error) {

	st := this.consensusState.GetState() // performs a copy
	cache := this.mempoolReactor.Mempool.GetCache()
	callee := &vm.Account{Address: cmn.Zero256}
	caller := &vm.Account{Address: cmn.Zero256}
	txCache := state.NewTxCache(cache)
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
func (this *transactor) BroadcastTx(tx types.Tx) (*Receipt, error) {
	err := this.mempoolReactor.BroadcastTx(tx)
	if err != nil {
		return nil, fmt.Errorf("Error broadcasting transaction: %v", err)
	}
	chainId := config.GetString("chain_id")
	txHash := types.TxID(chainId, tx)
	var createsContract uint8
	var contractAddr []byte
	// check if creates new contract
	if callTx, ok := tx.(*types.CallTx); ok {
		if len(callTx.Address) == 0 {
			createsContract = 1
			contractAddr = state.NewContractAddress(callTx.Input.Address, callTx.Input.Sequence)
		}
	}
	return &Receipt{txHash, createsContract, contractAddr}, nil
}

// Get all unconfirmed txs.
func (this *transactor) UnconfirmedTxs() (*UnconfirmedTxs, error) {
	transactions := this.mempoolReactor.Mempool.GetProposalTxs()
	return &UnconfirmedTxs{transactions}, nil
}

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

	pa := account.GenPrivAccountFromPrivKeyBytes(privKey)
	cache := this.mempoolReactor.Mempool.GetCache()
	acc := cache.GetAccount(pa.Address)
	var sequence int
	if acc == nil {
		sequence = 1
	} else {
		sequence = acc.Sequence + 1
	}
	txInput := &types.TxInput{
		Address:  pa.Address,
		Amount:   1,
		Sequence: sequence,
		PubKey:   pa.PubKey,
	}
	tx := &types.CallTx{
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

func (this *transactor) TransactNameReg(privKey []byte, name, data string, amount, fee int64) (*Receipt, error) {
	
	if len(privKey) != 64 {
		return nil, fmt.Errorf("Private key is not of the right length: %d\n", len(privKey))
	}

	pa := account.GenPrivAccountFromPrivKeyBytes(privKey)
	cache := this.mempoolReactor.Mempool.GetCache()
	acc := cache.GetAccount(pa.Address)
	var sequence int
	if acc == nil {
		sequence = 1
	} else {
		sequence = acc.Sequence + 1
	}
	tx := types.NewNameTxWithNonce(pa.PubKey, name, data, amount, fee, sequence)
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
	case *types.NameTx:
		nameTx := tx.(*types.NameTx)
		nameTx.Input.PubKey = privAccounts[0].PubKey
		nameTx.Input.Signature = privAccounts[0].Sign(config.GetString("chain_id"), nameTx)
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
		Nonce:       int64(acc.Sequence),
		StorageRoot: cmn.LeftPadWord256(acc.StorageRoot),
		Other:       acc.PubKey,
	}
}
