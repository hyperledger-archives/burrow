package errors

import fmt "fmt"

type Code uint32

const (
	ErrorCodeNone Code = iota
	ErrorCodeGeneric
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
	ErrorCodeUnresolvedSymbols
	ErrorCodeInvalidContractCode
	ErrorCodeNonExistentAccount
)

var codeStrings = map[Code]string{
	ErrorCodeNone:                   "no error",
	ErrorCodeUnknownAddress:         "unknown address",
	ErrorCodeInsufficientBalance:    "insufficient balance",
	ErrorCodeInvalidJumpDest:        "invalid jump dest",
	ErrorCodeInsufficientGas:        "insufficient gas",
	ErrorCodeMemoryOutOfBounds:      "memory out of bounds",
	ErrorCodeCodeOutOfBounds:        "code out of bounds",
	ErrorCodeInputOutOfBounds:       "input out of bounds",
	ErrorCodeReturnDataOutOfBounds:  " data out of bounds",
	ErrorCodeCallStackOverflow:      "call stack overflow",
	ErrorCodeCallStackUnderflow:     "call stack underflow",
	ErrorCodeDataStackOverflow:      "data stack overflow",
	ErrorCodeDataStackUnderflow:     "data stack underflow",
	ErrorCodeInvalidContract:        "invalid contract",
	ErrorCodePermissionDenied:       "permission denied",
	ErrorCodeNativeContractCodeCopy: "tried to copy native contract code",
	ErrorCodeExecutionAborted:       "execution aborted",
	ErrorCodeExecutionReverted:      "execution reverted",
	ErrorCodeNativeFunction:         "native function error",
	ErrorCodeEventPublish:           "event publish error",
	ErrorCodeInvalidString:          "invalid string",
	ErrorCodeEventMapping:           "event mapping error",
	ErrorCodeGeneric:                "generic error",
	ErrorCodeInvalidAddress:         "invalid address",
	ErrorCodeDuplicateAddress:       "duplicate address",
	ErrorCodeInsufficientFunds:      "insufficient funds",
	ErrorCodeOverpayment:            "overpayment",
	ErrorCodeZeroPayment:            "zero payment error",
	ErrorCodeInvalidSequence:        "invalid sequence number",
	ErrorCodeReservedAddress:        "address is reserved for SNative or internal use",
	ErrorCodeIllegalWrite:           "callee attempted to illegally modify state",
	ErrorCodeIntegerOverflow:        "integer overflow",
	ErrorCodeInvalidProposal:        "proposal is invalid",
	ErrorCodeExpiredProposal:        "proposal is expired since sequence number does not match",
	ErrorCodeProposalExecuted:       "proposal has already been executed",
	ErrorCodeNoInputPermission:      "account has no input permission",
	ErrorCodeInvalidBlockNumber:     "invalid block number",
	ErrorCodeBlockNumberOutOfRange:  "block number out of range",
	ErrorCodeAlreadyVoted:           "vote already registered for this address",
	ErrorCodeUnresolvedSymbols:      "code has unresolved symbols",
	ErrorCodeInvalidContractCode:    "contract being created with unexpected code",
	ErrorCodeNonExistentAccount:     "account does not exist",
}

func (c Code) ErrorCode() Code {
	return c
}

func (c Code) Uint32() uint32 {
	return uint32(c)
}

func (c Code) Error() string {
	return fmt.Sprintf("Error %d: %s", c, c.ErrorMessage())
}

func (c Code) ErrorMessage() string {
	str, ok := codeStrings[c]
	if !ok {
		return "Unknown error"
	}
	return str
}

func ErrorCode(err error) Code {
	exception := AsException(err)
	if exception == nil {
		return ErrorCodeNone
	}
	return exception.GetCode()
}
