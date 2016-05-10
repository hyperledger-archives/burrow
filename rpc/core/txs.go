package core

import (
	"fmt"

	acm "github.com/eris-ltd/eris-db/account"
	"github.com/eris-ltd/eris-db/evm"
	ctypes "github.com/eris-ltd/eris-db/rpc/core/types"
	"github.com/eris-ltd/eris-db/state"
	"github.com/eris-ltd/eris-db/txs"

	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
)

func toVMAccount(acc *acm.Account) *vm.Account {
	return &vm.Account{
		Address: LeftPadWord256(acc.Address),
		Balance: acc.Balance,
		Code:    acc.Code, // This is crazy.
		Nonce:   int64(acc.Sequence),
		Other:   acc.PubKey,
	}
}

//-----------------------------------------------------------------------------

// Run a contract's code on an isolated and unpersisted state
// Cannot be used to create new contracts
func Call(fromAddress, toAddress, data []byte) (*ctypes.ResultCall, error) {
	st := erisdbApp.GetState()
	cache := state.NewBlockCache(st)
	outAcc := cache.GetAccount(toAddress)
	if outAcc == nil {
		return nil, fmt.Errorf("Account %x does not exist", toAddress)
	}
	callee := toVMAccount(outAcc)
	caller := &vm.Account{Address: LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    st.GetGasLimit(),
	}

	vmach := vm.NewVM(txCache, params, caller.Address, nil)
	gas := st.GetGasLimit()
	ret, err := vmach.Call(caller, callee, callee.Code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultCall{Return: ret}, nil
}

// Run the given code on an isolated and unpersisted state
// Cannot be used to create new contracts
func CallCode(fromAddress, code, data []byte) (*ctypes.ResultCall, error) {

	st := erisdbApp.GetState()
	cache := erisdbApp.GetCheckCache()
	callee := &vm.Account{Address: LeftPadWord256(fromAddress)}
	caller := &vm.Account{Address: LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    st.GetGasLimit(),
	}

	vmach := vm.NewVM(txCache, params, caller.Address, nil)
	gas := st.GetGasLimit()
	ret, err := vmach.Call(caller, callee, code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultCall{Return: ret}, nil
}

//-----------------------------------------------------------------------------

func SignTx(tx txs.Tx, privAccounts []*acm.PrivAccount) (*ctypes.ResultSignTx, error) {
	// more checks?

	for i, privAccount := range privAccounts {
		if privAccount == nil || privAccount.PrivKey == nil {
			return nil, fmt.Errorf("Invalid (empty) privAccount @%v", i)
		}
	}
	switch tx.(type) {
	case *txs.SendTx:
		sendTx := tx.(*txs.SendTx)
		for i, input := range sendTx.Inputs {
			input.PubKey = privAccounts[i].PubKey
			input.Signature = privAccounts[i].Sign(config.GetString("erisdb.chain_id"), sendTx)
		}
	case *txs.CallTx:
		callTx := tx.(*txs.CallTx)
		callTx.Input.PubKey = privAccounts[0].PubKey
		callTx.Input.Signature = privAccounts[0].Sign(config.GetString("erisdb.chain_id"), callTx)
	case *txs.BondTx:
		bondTx := tx.(*txs.BondTx)
		// the first privaccount corresponds to the BondTx pub key.
		// the rest to the inputs
		bondTx.Signature = privAccounts[0].Sign(config.GetString("erisdb.chain_id"), bondTx).(crypto.SignatureEd25519)
		for i, input := range bondTx.Inputs {
			input.PubKey = privAccounts[i+1].PubKey
			input.Signature = privAccounts[i+1].Sign(config.GetString("erisdb.chain_id"), bondTx)
		}
	case *txs.UnbondTx:
		unbondTx := tx.(*txs.UnbondTx)
		unbondTx.Signature = privAccounts[0].Sign(config.GetString("erisdb.chain_id"), unbondTx).(crypto.SignatureEd25519)
	case *txs.RebondTx:
		rebondTx := tx.(*txs.RebondTx)
		rebondTx.Signature = privAccounts[0].Sign(config.GetString("erisdb.chain_id"), rebondTx).(crypto.SignatureEd25519)
	}
	return &ctypes.ResultSignTx{tx}, nil
}
