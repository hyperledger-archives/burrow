package payload

import (
	"errors"
	"fmt"
)

var (
	ErrTxInvalidAddress    = errors.New("error invalid address")
	ErrTxDuplicateAddress  = errors.New("error duplicate address")
	ErrTxInvalidAmount     = errors.New("error invalid amount")
	ErrTxInsufficientFunds = errors.New("error insufficient funds")
	ErrTxUnknownPubKey     = errors.New("error unknown pubkey")
	ErrTxInvalidPubKey     = errors.New("error invalid pubkey")
	ErrTxInvalidSignature  = errors.New("error invalid signature")
)

type ErrTxInvalidString struct {
	Msg string
}

func (e ErrTxInvalidString) Error() string {
	return e.Msg
}

type ErrTxInvalidSequence struct {
	Got      uint64
	Expected uint64
}

func (e ErrTxInvalidSequence) Error() string {
	return fmt.Sprintf("Error invalid sequence. Got %d, expected %d", e.Got, e.Expected)
}
