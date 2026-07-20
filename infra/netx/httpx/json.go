package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/duakc/mt"
)

// JSONRequest performs an HTTP request whose body and response are JSON.
//
// Acceptable response status codes are determined by policy.Accept
// (default: any code < 400). Pass a zero-valued RespPolicy to use defaults,
// or set RespPolicy.AcceptCode to a custom predicate to override.
//
// The caller owns the returned *http.Response and is responsible for closing
// its Body when it is non-nil — including on error paths, where the response
// may still be non-nil (e.g. status / Content-Type / decode errors).
func JSONRequest[O any](
	ctx context.Context, do HTTPRequester, req ReqConfig, policy RespPolicy,
) (O, *http.Response, error) {
	if req.ExtendHeader == nil {
		req.ExtendHeader = make(http.Header)
	}

	if req.ExtendHeader.Get("Content-Type") == "" {
		req.ExtendHeader.Set("Content-Type", "application/json; charset=utf-8")
	}

	request, err := req.ToRequestContext(ctx)
	if err != nil {
		return mt.Zero[O](), nil, fmt.Errorf("toRequest: %w", err)
	}

	response, err := do.Do(request)
	if err != nil {
		return mt.Zero[O](), nil, NewBaseResponseError(err, req.Method, "")
	}

	if err := policy.AcceptResponse(response); err != nil {
		return mt.Zero[O](), response, err
	}

	if ct := response.Header.Get("Content-Type"); !IsJsonContentType(ct) {
		return mt.Zero[O](), response, fmt.Errorf("not a JSON response: Content-Type: %s", ct)
	}

	decoder := json.NewDecoder(response.Body)
	var output O
	err = decoder.Decode(&output)
	if err != nil {
		return mt.Zero[O](), response, fmt.Errorf("decode JSON: %w", err)
	}
	return output, response, nil
}
