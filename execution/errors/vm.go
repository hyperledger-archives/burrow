package errors

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
)

type PermissionDenied struct {
	Address crypto.Address
	Perm    permission.PermFlag
}

func (err PermissionDenied) ErrorCode() *Code {
	return Codes.PermissionDenied
}

func (err PermissionDenied) Error() string {
	return fmt.Sprintf("Account/contract %v does not have permission %v", err.Address, err.Perm)
}

type NestedCallError struct {
	CodedError
	Caller     crypto.Address
	Callee     crypto.Address
	StackDepth uint64
}

func (err NestedCallError) Error() string {
	return fmt.Sprintf("error in nested call at depth %v: %s (callee) -> %s (caller): %v",
		err.StackDepth, err.Callee, err.Caller, err.CodedError)
}

type CallError struct {
	// The error from the original call which defines the overall error code
	CodedError
	// Errors from nested sub-calls of the original call that may have also occurred
	NestedErrors []NestedCallError
}

func (err CallError) Error() string {
	return fmt.Sprintf("Call error: %v (and %d nested sub-call errors)", err.CodedError, len(err.NestedErrors))
}
