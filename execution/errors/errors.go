package errors

import (
	"encoding/json"
	"fmt"
)

type CodedError interface {
	error
	ErrorCode() Code
}

type Code uint32

const (
	ErrorCodeGeneric Code = iota
	ErrorCodeUnknownAddress
	ErrorCodeInsufficientBalance
	ErrorCodeInvalidJumpDest
	ErrorCodeInsufficientGas
	ErrorCodeMemoryOutOfBounds
	ErrorCodeCodeOutOfBounds
	ErrorCodeInputOutOfBounds
	ErrorCodeReturnDataOutOfBounds
	ErrorCodeCallStackOverflow
	ErrorCodeCallStackUnderflow
	ErrorCodeDataStackOverflow
	ErrorCodeDataStackUnderflow
	ErrorCodeInvalidContract
	ErrorCodeNativeContractCodeCopy
	ErrorCodeExecutionAborted
	ErrorCodeExecutionReverted
	ErrorCodePermissionDenied
	ErrorCodeNativeFunction
	ErrorCodeEventPublish
)

func (c Code) ErrorCode() Code {
	return c
}

func (c Code) Error() string {
	switch c {
	case ErrorCodeUnknownAddress:
		return "Unknown address"
	case ErrorCodeInsufficientBalance:
		return "Insufficient balance"
	case ErrorCodeInvalidJumpDest:
		return "Invalid jump dest"
	case ErrorCodeInsufficientGas:
		return "Insufficient gas"
	case ErrorCodeMemoryOutOfBounds:
		return "Memory out of bounds"
	case ErrorCodeCodeOutOfBounds:
		return "Code out of bounds"
	case ErrorCodeInputOutOfBounds:
		return "Input out of bounds"
	case ErrorCodeReturnDataOutOfBounds:
		return "Return data out of bounds"
	case ErrorCodeCallStackOverflow:
		return "Call stack overflow"
	case ErrorCodeCallStackUnderflow:
		return "Call stack underflow"
	case ErrorCodeDataStackOverflow:
		return "Data stack overflow"
	case ErrorCodeDataStackUnderflow:
		return "Data stack underflow"
	case ErrorCodeInvalidContract:
		return "Invalid contract"
	case ErrorCodeNativeContractCodeCopy:
		return "Tried to copy native contract code"
	case ErrorCodeExecutionAborted:
		return "Execution aborted"
	case ErrorCodeExecutionReverted:
		return "Execution reverted"
	case ErrorCodeNativeFunction:
		return "Native function error"
	case ErrorCodeEventPublish:
		return "Event publish error"
	case ErrorCodeGeneric:
		return "Generic error"
	default:
		return "Unknown error"
	}
}

func NewCodedError(errorCode Code, exception string) *Exception {
	if exception == "" {
		return nil
	}
	return &Exception{
		Code: &ErrorCode{
			Code: uint32(errorCode),
		},
		Exception: exception,
	}
}

// Wraps any error as a Exception
func AsCodedError(err error) *Exception {
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case *Exception:
		return e
	case CodedError:
		return NewCodedError(e.ErrorCode(), e.Error())
	default:
		return NewCodedError(ErrorCodeGeneric, err.Error())
	}
}

func Wrap(err CodedError, message string) *Exception {
	return NewCodedError(err.ErrorCode(), message+": "+err.Error())
}

func Errorf(format string, a ...interface{}) CodedError {
	return ErrorCodef(ErrorCodeGeneric, format, a...)
}

func ErrorCodef(errorCode Code, format string, a ...interface{}) CodedError {
	return NewCodedError(errorCode, fmt.Sprintf(format, a...))
}

func (e *Exception) AsError() error {
	// We need to return a bare untyped error here so that err == nil downstream
	if e == nil {
		return nil
	}
	return e
}

func (e *Exception) ErrorCode() Code {
	return Code(e.GetCode().Code)
}

func (e *Exception) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("Error %v: %s", e.Code.Code, e.Exception)
}

func NewErrorCode(code Code) *ErrorCode {
	return &ErrorCode{
		Code: uint32(code),
	}
}

func (e ErrorCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Code)
}

func (e *ErrorCode) UnmarshalJSON(bs []byte) error {
	return json.Unmarshal(bs, &(e.Code))
}
