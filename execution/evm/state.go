package evm

import (
	"github.com/hyperledger/burrow/account"
)

type State interface {
	account.Creator
	account.Updater
	account.Storage
}
