package httpx

import (
	"fmt"
	"net/http"
)

type RespPolicy struct {
	AcceptCode func(code int) bool
}

func (rp RespPolicy) AcceptResponse(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}

	if !rp.acceptCodes(resp.StatusCode) {
		return &BadStatusCodeError{Got: resp.StatusCode}
	}

	return nil
}

func (rp RespPolicy) acceptCodes(code int) bool {
	if rp.AcceptCode == nil {
		return code < 400
	}
	return rp.AcceptCode(code)
}

type StatusCodeResponseWriter struct {
	http.ResponseWriter

	statusCode int
}

func (w *StatusCodeResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *StatusCodeResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
