package httpx

import (
	"fmt"
)

// BadStatusCodeError is returned when a response status code is not accepted
// by the caller's RespPolicy.AcceptCode predicate.
type BadStatusCodeError struct {
	Got int `json:"got"`
}

func (E *BadStatusCodeError) Error() string {
	return fmt.Sprintf("unacceptable status code: %d", E.Got)
}

type BaseResponseError struct {
	Err error `json:"err"`

	Method  string `json:"method"`
	Message string `json:"message"` // optional
}

func NewBaseResponseError(err error, method string, message string) *BaseResponseError {
	return &BaseResponseError{
		Method:  method,
		Err:     err,
		Message: message,
	}
}

func (E *BaseResponseError) Error() string {
	formatMessage := "method=%s"
	if E.Message != "" {
		formatMessage += ": " + E.Message
	}
	return fmt.Sprintf(formatMessage+": %s", E.Method, E.Err.Error())
}

func (E *BaseResponseError) Unwrap() error {
	return E.Err
}
