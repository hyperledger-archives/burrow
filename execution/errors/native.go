package errors

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

type LacksSNativePermission struct {
	Address crypto.Address
	SNative string
}

var _ CodedError = &LacksSNativePermission{}

func (e *LacksSNativePermission) ErrorMessage() string {
	return fmt.Sprintf("account %s does not have SNative function call permission: %s", e.Address, e.SNative)
}

func (e *LacksSNativePermission) Error() string {
	return e.ErrorMessage()
}

func (e *LacksSNativePermission) ErrorCode() Code {
	return ErrorCodeNativeFunction
}
