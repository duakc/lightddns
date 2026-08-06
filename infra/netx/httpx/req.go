package httpx

import (
	"net/http"

	"go.uber.org/zap"
)

type HTTPRequester interface {
	// Do send an HTTP request defined in the r
	// Response can not be nil if err is nil
	Do(r *http.Request) (*http.Response, error)
}

type TokenClient struct {
	HTTPRequester

	Token string
}

func (tc *TokenClient) Do(r *http.Request) (*http.Response, error) {
	if tc.Token != "" {
		r.Header.Set(HeaderAuthorization, "Bearer "+tc.Token)
	}

	return tc.HTTPRequester.Do(r)
}

type HeaderClient struct {
	HTTPRequester

	Headers http.Header
}

func (hc *HeaderClient) Do(r *http.Request) (*http.Response, error) {
	if len(hc.Headers) == 0 {
		return hc.HTTPRequester.Do(r)
	}
	ExtendHeaders(r.Header, hc.Headers)
	return hc.HTTPRequester.Do(r)
}

type DebugClient struct {
	HTTPRequester

	Logger *zap.Logger
}

func (dc *DebugClient) Do(r *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	if dc.Logger != nil && dc.Logger.Level().Enabled(zap.DebugLevel) {
		defer NewHTTPRequestRecorder(dc.Logger, r, &resp, &err)
	}
	return resp, err
}
