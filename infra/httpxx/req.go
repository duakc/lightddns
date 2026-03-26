package httpxx

import "net/http"

type HTTPRequester interface {

	// Do send an HTTP request defined in the r
	// Response can not be nil if err is nil
	Do(r *http.Request) (*http.Response, error)
}

type ValidClient struct {
	HTTPRequester
}

func (vc *ValidClient) Do(r *http.Request) (*http.Response, error) {
	rr, e := vc.HTTPRequester.Do(r)
	if e == nil && r == nil {
		panic("bad implementation of HTTPRequester")
	}
	return rr, e
}

type TokenClient struct {
	HTTPRequester
	Token string
}

func (tc *TokenClient) Do(r *http.Request) (*http.Response, error) {
	if tc.Token != "" {
		r.Header.Add("Authorization", "Bearer "+tc.Token)
	}

	return tc.HTTPRequester.Do(r)
}
