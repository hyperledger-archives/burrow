package errors

import (
	"bytes"
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission/types"
)

type PermissionDenied struct {
	Perm types.PermFlag
}

func (err PermissionDenied) ErrorCode() Code {
	return ErrorCodePermissionDenied
}

func (err PermissionDenied) Error() string {
	return fmt.Sprintf("Contract does not have permission to %v", err.Perm)
}

type NestedCall struct {
	NestedError CodedError
	Caller      crypto.Address
	Callee      crypto.Address
	StackDepth  uint64
}

func (err NestedCall) ErrorCode() Code {
	return err.NestedError.ErrorCode()
}

func (err NestedCall) Error() string {
	return fmt.Sprintf("error in nested call at depth %v: %s (callee) -> %s (caller): %v",
		err.StackDepth, err.Callee, err.Caller, err.NestedError)
}

type Call struct {
	CallError    CodedError
	NestedErrors []NestedCall
}

func (err Call) ErrorCode() Code {
	return err.CallError.ErrorCode()
}

func (err Call) Error() string {
	buf := new(bytes.Buffer)
	buf.WriteString("call error: ")
	buf.WriteString(err.CallError.Error())
	if len(err.NestedErrors) > 0 {
		buf.WriteString(", nested call errors:\n")
		for _, nestedErr := range err.NestedErrors {
			buf.WriteString(nestedErr.Error())
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}
