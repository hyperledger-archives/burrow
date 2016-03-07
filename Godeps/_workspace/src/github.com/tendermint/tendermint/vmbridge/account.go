package vmbridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

var acc *Account
var accI vm.Account = acc

// Implements github.com/ethereum/go-ethereum/core/vm.Account
type Account struct {
}

func (acc *Account) SubBalance(amount *big.Int) {
}

func (acc *Account) AddBalance(amount *big.Int) {
}

func (acc *Account) SetBalance(*big.Int) {
}

func (acc *Account) SetNonce(uint64) {
}

func (acc *Account) Balance() *big.Int {
	return nil
}

func (acc *Account) Address() common.Address {
	return common.Address{}
}

func (acc *Account) ReturnGas(*big.Int, *big.Int) {
}

func (acc *Account) SetCode([]byte) {
}

func (acc *Account) EachStorage(cb func(key, value []byte)) {
}

func (acc *Account) Value() *big.Int {
	return nil
}
