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

type ResponseError struct {
	Method  string `json:"method"`
	URL     string `json:"url"`
	Err     error  `json:"err"`
	Message string `json:"message"` // optional
}

func NewResponseError(method, url string, err error) *ResponseError {
	return &ResponseError{
		Method: method,
		URL:    url,
		Err:    err,
	}
}

func (E *ResponseError) Error() string {
	formatMessage := "an error occurred while requesting %s: method=%s"
	if E.Message != "" {
		formatMessage += ": " + E.Message
	}
	return fmt.Sprintf(formatMessage+": %s", E.URL, E.Method, E.Err.Error())
}
