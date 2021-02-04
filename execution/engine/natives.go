package engine

import (
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
)

type Native interface {
	Callable
	SetExternals(externals Dispatcher)
	ContractMeta() []*acm.ContractMeta
	FullName() string
	Address() crypto.Address
}

type Natives interface {
	ExternalDispatcher
	GetByAddress(address crypto.Address) Native
}
