package state

import (
	acm "github.com/eris-ltd/eris-db/account"
	. "github.com/tendermint/go-common"
	"github.com/eris-ltd/eris-db/vm"
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
