package httpxx

import "net/http"

type HTTPRequester interface {
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
