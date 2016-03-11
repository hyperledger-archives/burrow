package vmbridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	//"github.com/ethereum/go-ethereum/core/vm"

	mintvm "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/vm"
)

// Implements github.com/ethereum/go-ethereum/core/vm.Account
type Account struct {
	acc *mintvm.Account
}

func NewAccount(acc *mintvm.Account) *Account {
	return &Account{
		acc: acc,
	}
}

func (acc *Account) SubBalance(amount *big.Int) {
	acc.acc.Balance -= amount.Int64() // XXX: overflow
}

func (acc *Account) AddBalance(amount *big.Int) {
	acc.acc.Balance += amount.Int64() // XXX: overflow
}

func (acc *Account) SetBalance(balance *big.Int) {
	acc.acc.Balance = balance.Int64() // XXX: overflow
}

func (acc *Account) SetNonce(nonce uint64) {
	acc.acc.Nonce = int64(nonce) // XXX: overflow
}

func (acc *Account) Balance() *big.Int {
	return big.NewInt(acc.acc.Balance)
}

func (acc *Account) Address() common.Address {
	return common.BytesToAddress(acc.acc.Address.Postfix(20))
}

func (acc *Account) ReturnGas(*big.Int, *big.Int) {
	// in eth: implemented for vm.Contract, but not state.StateObject

	/*
		// ReturnGas adds the given gas back to itself.
		func (c *Contract) ReturnGas(gas, price *big.Int) {
				// Return the gas to the context
					c.Gas.Add(c.Gas, gas)
						c.UsedGas.Sub(c.UsedGas, gas)
					}
	*/

}

func (acc *Account) SetCode(code []byte) {
	acc.acc.Code = code
}

func (acc *Account) EachStorage(cb func(key, value []byte)) {
	// XXX: see no reason to implement this ...
}

func (acc *Account) Value() *big.Int {
	// in eth, only implemented for vm.Contract, panics for state.StateObject
	return big.NewInt(acc.acc.Balance)
}
