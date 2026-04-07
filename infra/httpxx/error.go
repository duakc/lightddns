package httpxx

import (
	"fmt"
	"net/http"
)

type BadStatusCodeError struct {
	Excepted int `json:"excepted"`
	Got      int `json:"got"`
}

func (E *BadStatusCodeError) Error() string {
	if E.Excepted == 0 {
		E.Excepted = http.StatusOK
	}
	return fmt.Sprintf("bad status code: excepted %d, got %d", E.Excepted, E.Got)
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
