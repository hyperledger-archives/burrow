package state

import (
	. "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/go-common"
	acm "github.com/eris-ltd/eris-db/account"
	"github.com/eris-ltd/eris-db/evm"
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
