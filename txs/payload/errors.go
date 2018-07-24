package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/execution/errors"
)

type ErrTxInvalidSequence struct {
	Input   *TxInput
	Account acm.Account
}

func (e ErrTxInvalidSequence) Error() string {
	return fmt.Sprintf("Error invalid sequence in input %v: input has sequence %d, but account has sequence %d, "+
		"so expected input to have sequence %d", e.Input, e.Input.Sequence, e.Account.Sequence(), e.Account.Sequence()+1)
}

func (e ErrTxInvalidSequence) ErrorCode() errors.Code {
	return errors.ErrorCodeInvalidSequence
}
