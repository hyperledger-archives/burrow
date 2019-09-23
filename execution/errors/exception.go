package errors

import "fmt"

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
		return NewException(e.ErrorCode(), e.ErrorMessage())
	default:
		return NewException(ErrorCodeGeneric, err.Error())
	}
}

func Wrapf(err error, format string, a ...interface{}) *Exception {
	ex := AsException(err)
	return NewException(ex.Code, fmt.Sprintf(format, a...))
}

func Wrap(err error, message string) *Exception {
	ex := AsException(err)
	return NewException(ex.Code, message+": "+ex.Exception)
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
	return fmt.Sprintf("error %d - %s: %s", e.Code, e.Code.ErrorMessage(), e.Exception)
}

func (e *Exception) String() string {
	return e.Error()
}

func (e *Exception) ErrorMessage() string {
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
