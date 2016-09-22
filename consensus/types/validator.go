package types

import (
	"github.com/tendermint/go-wire"
)

var _ = wire.RegisterInterface(
	struct{ Validator }{},
	wire.ConcreteType{&TendermintValidator{}, byte(0x01)},
)

type Validator interface {
	AssertIsValidator()
	Address() []byte
}
