package httpx

import "net/http"

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
		r.Header.Add("Authorization", "Bearer "+tc.Token)
	}

	return tc.HTTPRequester.Do(r)
}
