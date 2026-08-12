package transports

import (
	"errors"
	"fmt"

	mDns "github.com/miekg/dns"
)

type RcodeError struct {
	Code     int
	Excepted int
}

func (e *RcodeError) Error() string {
	return fmt.Sprintf("bad rcode: %s, excepted: %s", mDns.RcodeToString[e.Code], mDns.RcodeToString[e.Excepted])
}

// RetryableError marks a transport error that the resolver may retry with the
// same DNS message.
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// Retryable wraps err as retryable for resolver-level retry handling.
func Retryable(err error) error {
	if err == nil || IsRetryable(err) {
		return err
	}
	return &RetryableError{Err: err}
}

// IsRetryable reports whether err has been marked for resolver-level retry.
func IsRetryable(err error) bool {
	var retryableError *RetryableError
	return errors.As(err, &retryableError)
}
