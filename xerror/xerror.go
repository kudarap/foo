package xerror

import "fmt"

type Error string

func (e Error) Error() string { return string(e) }

// X extends error to coded error and return as XError.
func (e Error) X(err error) XError { return NewXError(err, string(e)) }

// XError represents extended error details.
type XError struct {
	Err  error
	Code string
}

func NewXError(err error, code string) XError {
	return XError{Err: err, Code: code}
}

func (e XError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Err) }
