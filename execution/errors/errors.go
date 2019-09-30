package errors

type CodedError interface {
	error
	ErrorCode() Code
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
