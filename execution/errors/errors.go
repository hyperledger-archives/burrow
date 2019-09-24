package errors

type CodedError interface {
	error
	ErrorCode() *Coding
	// The error message excluding the code
	ErrorMessage() string
}

// Error sinks are useful for recording errors but continuing with computation. Implementations may choose to just store
// the first error pushed and ignore subsequent ones or they may record an error trace. Pushing a nil error should have
// no effects.
type Sink interface {
	// Push an error to the error. If a nil error is passed then that value should not be pushed. Returns true iff error
	// is non nil.
	PushError(error) bool
}

type Source interface {
	// Returns the an error if errors occurred some execution or nil if none occurred
	Error() error
}

var Code = Codes{
	None:                   description("none"),
	UnknownAddress:         description("unknown address"),
	InsufficientBalance:    description("insufficient balance"),
	InvalidJumpDest:        description("invalid jump dest"),
	InsufficientGas:        description("insufficient gas"),
	MemoryOutOfBounds:      description("memory out of bounds"),
	CodeOutOfBounds:        description("code out of bounds"),
	InputOutOfBounds:       description("input out of bounds"),
	ReturnDataOutOfBounds:  description("data out of bounds"),
	CallStackOverflow:      description("call stack overflow"),
	CallStackUnderflow:     description("call stack underflow"),
	DataStackOverflow:      description("data stack overflow"),
	DataStackUnderflow:     description("data stack underflow"),
	InvalidContract:        description("invalid contract"),
	PermissionDenied:       description("permission denied"),
	NativeContractCodeCopy: description("tried to copy native contract code"),
	ExecutionAborted:       description("execution aborted"),
	ExecutionReverted:      description("execution reverted"),
	NativeFunction:         description("native function error"),
	EventPublish:           description("event publish error"),
	InvalidString:          description("invalid string"),
	EventMapping:           description("event mapping error"),
	Generic:                description("generic error"),
	InvalidAddress:         description("invalid address"),
	DuplicateAddress:       description("duplicate address"),
	InsufficientFunds:      description("insufficient funds"),
	Overpayment:            description("overpayment"),
	ZeroPayment:            description("zero payment error"),
	InvalidSequence:        description("invalid sequence number"),
	ReservedAddress:        description("address is reserved for SNative or internal use"),
	IllegalWrite:           description("callee attempted to illegally modify state"),
	IntegerOverflow:        description("integer overflow"),
	InvalidProposal:        description("proposal is invalid"),
	ExpiredProposal:        description("proposal is expired since sequence number does not match"),
	ProposalExecuted:       description("proposal has already been executed"),
	NoInputPermission:      description("account has no input permission"),
	InvalidBlockNumber:     description("invalid block number"),
	BlockNumberOutOfRange:  description("block number out of range"),
	AlreadyVoted:           description("vote already registered for this address"),
	UnresolvedSymbols:      description("code has unresolved symbols"),
	InvalidContractCode:    description("contract being created with unexpected code"),
	NonExistentAccount:     description("account does not exist"),
}

func init() {
	err := Code.init()
	if err != nil {
		panic(err)
	}
}
