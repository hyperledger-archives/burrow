package errors

import "fmt"

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
	ErrorCodeInvalidString
	ErrorCodeEventMapping
	ErrorCodeInvalidAddress
	ErrorCodeDuplicateAddress
	ErrorCodeInsufficientFunds
	ErrorCodeOverpayment
	ErrorCodeZeroPayment
	ErrorCodeInvalidSequence
	ErrorCodeReservedAddress
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
	case ErrorCodePermissionDenied:
		return "Permission denied"
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
	case ErrorCodeInvalidString:
		return "Invalid string"
	case ErrorCodeEventMapping:
		return "Event mapping error"
	case ErrorCodeGeneric:
		return "Generic error"
	case ErrorCodeInvalidAddress:
		return "Invalid address"
	case ErrorCodeDuplicateAddress:
		return "Duplicate address"
	case ErrorCodeInsufficientFunds:
		return "Insufficient funds"
	case ErrorCodeOverpayment:
		return "Overpayment"
	case ErrorCodeZeroPayment:
		return "Zero payment error"
	case ErrorCodeInvalidSequence:
		return "Invalid sequence number"
	case ErrorCodeReservedAddress:
		return "Address is reserved for SNative or internal use"
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
		return NewException(e.ErrorCode(), e.Error())
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

func (e *Exception) String() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Exception)
}

func (e *Exception) Error() string {
	if e == nil {
		return ""
	}
	return e.Exception
}
