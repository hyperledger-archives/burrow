package errors

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

type LacksSNativePermission struct {
	Address crypto.Address
	SNative string
}

func (e LacksSNativePermission) Error() string {
	return fmt.Sprintf("account %s does not have SNative function call permission: %s", e.Address, e.SNative)
}

func (e LacksSNativePermission) ErrorCode() ErrorCode {
	return ErrorCodeNativeFunction
}
