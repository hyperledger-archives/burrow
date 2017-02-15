package state

import (
	acm "github.com/eris-ltd/eris-db/account"
	"github.com/eris-ltd/eris-db/manager/eris-mint/evm"
	. "github.com/eris-ltd/eris-db/word256"
)

type AccountGetter interface {
	GetAccount(addr []byte) *acm.Account
}

type VMAccountState interface {
	GetAccount(addr Word256) *vm.Account
	UpdateAccount(acc *vm.Account)
	RemoveAccount(acc *vm.Account)
	CreateAccount(creator *vm.Account) *vm.Account
}
