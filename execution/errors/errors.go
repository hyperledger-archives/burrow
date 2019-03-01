package errors

import (
	"fmt"
)

type CodedError interface {
	error
	ErrorCode() Code
	String() string
}

type Provider interface {
	// Returns the an error if errors occurred some execution or nil if none occurred
	Error() CodedError
}

type Sink interface {
	PushError(error)
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
	ErrorCodeInvalidString
	ErrorCodeEventMapping
	ErrorCodeInvalidAddress
	ErrorCodeDuplicateAddress
	ErrorCodeInsufficientFunds
	ErrorCodeOverpayment
	ErrorCodeZeroPayment
	ErrorCodeInvalidSequence
	ErrorCodeReservedAddress
	ErrorCodeIllegalWrite
	ErrorCodeIntegerOverflow
	ErrorCodeInvalidProposal
	ErrorCodeExpiredProposal
	ErrorCodeProposalExecuted
	ErrorCodeNoInputPermission
	ErrorCodeInvalidBlockNumber
	ErrorCodeBlockNumberOutOfRange
	ErrorCodeAlreadyVoted
)

func (c Code) ErrorCode() Code {
	return c
}

func (c Code) Uint32() uint32 {
	return uint32(c)
}

func (c Code) Error() string {
	return fmt.Sprintf("Error %d: %s", c, c.String())
}

func (c Code) String() string {
	switch c {
	case ErrorCodeUnknownAddress:
		return "unknown address"
	case ErrorCodeInsufficientBalance:
		return "insufficient balance"
	case ErrorCodeInvalidJumpDest:
		return "invalid jump dest"
	case ErrorCodeInsufficientGas:
		return "insufficient gas"
	case ErrorCodeMemoryOutOfBounds:
		return "memory out of bounds"
	case ErrorCodeCodeOutOfBounds:
		return "code out of bounds"
	case ErrorCodeInputOutOfBounds:
		return "input out of bounds"
	case ErrorCodeReturnDataOutOfBounds:
		return "return data out of bounds"
	case ErrorCodeCallStackOverflow:
		return "call stack overflow"
	case ErrorCodeCallStackUnderflow:
		return "call stack underflow"
	case ErrorCodeDataStackOverflow:
		return "data stack overflow"
	case ErrorCodeDataStackUnderflow:
		return "data stack underflow"
	case ErrorCodeInvalidContract:
		return "invalid contract"
	case ErrorCodePermissionDenied:
		return "permission denied"
	case ErrorCodeNativeContractCodeCopy:
		return "tried to copy native contract code"
	case ErrorCodeExecutionAborted:
		return "execution aborted"
	case ErrorCodeExecutionReverted:
		return "execution reverted"
	case ErrorCodeNativeFunction:
		return "native function error"
	case ErrorCodeEventPublish:
		return "event publish error"
	case ErrorCodeInvalidString:
		return "invalid string"
	case ErrorCodeEventMapping:
		return "event mapping error"
	case ErrorCodeGeneric:
		return "generic error"
	case ErrorCodeInvalidAddress:
		return "invalid address"
	case ErrorCodeDuplicateAddress:
		return "duplicate address"
	case ErrorCodeInsufficientFunds:
		return "insufficient funds"
	case ErrorCodeOverpayment:
		return "overpayment"
	case ErrorCodeZeroPayment:
		return "zero payment error"
	case ErrorCodeInvalidSequence:
		return "invalid sequence number"
	case ErrorCodeReservedAddress:
		return "address is reserved for SNative or internal use"
	case ErrorCodeIllegalWrite:
		return "callee attempted to illegally modify state"
	case ErrorCodeIntegerOverflow:
		return "integer overflow"
	case ErrorCodeInvalidProposal:
		return "proposal is invalid"
	case ErrorCodeExpiredProposal:
		return "proposal is expired since sequence number does not match"
	case ErrorCodeProposalExecuted:
		return "proposal has already been executed"
	case ErrorCodeNoInputPermission:
		return "account has no input permission"
	case ErrorCodeInvalidBlockNumber:
		return "invalid block number"
	case ErrorCodeBlockNumberOutOfRange:
		return "block number out of range"
	case ErrorCodeAlreadyVoted:
		return "vote already registered for this address"
	default:
		return "Unknown error"
	}
}

func NewException(errorCode Code, exception string) *Exception {
	if exception == "" {
		return nil
	}
	return &Exception{
		Code:      errorCode,
		Exception: exception,
	}
}

// Wraps any error as a Exception
func AsException(err error) *Exception {
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case *Exception:
		return e
	case CodedError:
		return NewException(e.ErrorCode(), e.String())
	default:
		return NewException(ErrorCodeGeneric, err.Error())
	}
}

func Wrap(err error, message string) *Exception {
	ex := AsException(err)
	return NewException(ex.ErrorCode(), message+": "+ex.Error())
}

func Errorf(format string, a ...interface{}) *Exception {
	return ErrorCodef(ErrorCodeGeneric, format, a...)
}

func ErrorCodef(errorCode Code, format string, a ...interface{}) *Exception {
	return NewException(errorCode, fmt.Sprintf(format, a...))
}

func (e *Exception) AsError() error {
	// We need to return a bare untyped error here so that err == nil downstream
	if e == nil {
		return nil
	}
	return e
}

func (e *Exception) ErrorCode() Code {
	return e.Code
}

func (e *Exception) Error() string {
	return fmt.Sprintf("error %d - %s: %s", e.Code, e.Code.String(), e.Exception)
}

func (e *Exception) String() string {
	if e == nil {
		return ""
	}
	return e.Exception
}

func (e *Exception) Equal(ce CodedError) bool {
	ex := AsException(ce)
	if e == nil || ex == nil {
		return e == nil && ex == nil
	}
	return e.Code == ex.Code && e.Exception == ex.Exception
}

type singleError struct {
	CodedError
}

func FirstOnly() *singleError {
	return &singleError{}
}

func (se *singleError) PushError(err error) {
	if se.CodedError == nil {
		// Do our nil dance
		ex := AsException(err)
		if ex != nil {
			se.CodedError = ex
		}
	}
}

func (se *singleError) Error() CodedError {
	return se.CodedError
}

func (se *singleError) Reset() {
	se.CodedError = nil
}
