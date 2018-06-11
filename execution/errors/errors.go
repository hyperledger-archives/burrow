package errors

import (
	"fmt"
)

type CodedError interface {
	error
	ErrorCode() ErrorCode
}

type ErrorCode int8

const (
	ErrorCodeGeneric ErrorCode = iota
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
)

func (ec ErrorCode) ErrorCode() ErrorCode {
	return ec
}

func (ec ErrorCode) Error() string {
	switch ec {
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
	default:
		return "Generic error"
	}
}

// Exception provides a serialisable coded error for the VM
type Exception struct {
	Code      ErrorCode
	Exception string
}

func NewCodedError(errorCode ErrorCode, exception string) *Exception {
	if exception == "" {
		return nil
	}
	return &Exception{
		Code:      errorCode,
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

func ErrorCodef(errorCode ErrorCode, format string, a ...interface{}) CodedError {
	return NewCodedError(errorCode, fmt.Sprintf(format, a...))
}

func (e *Exception) AsError() error {
	// thanks go, you dick
	if e == nil {
		return nil
	}
	return e
}

func (e *Exception) ErrorCode() ErrorCode {
	return e.Code
}

func (e *Exception) String() string {
	return e.Error()
}

func (e *Exception) Error() string {
	return fmt.Sprintf("VM Error %v: %s", e.Code, e.Exception)
}
