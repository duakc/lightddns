package httpxx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/duakc/mt"
)

// JSONRequest performs an HTTP request whose body and response are JSON.
//
// Acceptable response status codes are determined by req.AcceptStatus
// (default: any code < 400). Set ReqConfig.AcceptStatus to a custom predicate
// to override.
//
// The caller owns the returned *http.Response and is responsible for closing
// its Body when it is non-nil — including on error paths, where the response
// may still be non-nil (e.g. status / Content-Type / decode errors).
func JSONRequest[O any](ctx context.Context, do HTTPRequester, req ReqConfig, input any) (O, *http.Response, error) {
	if input != nil {
		// Stamp Content-Type=application/json on a cloned header so the
		// caller's ReqConfig stays untouched. ToRequestContext then picks
		// the right marshal path: json.Marshaler if input implements it,
		// json.Marshal as the fallback (handles map[string]any, structs,
		// slices, primitives).
		header := req.ExtendHeader.Clone()
		if header == nil {
			header = make(http.Header)
		}
		header.Set("Content-Type", "application/json")
		req.ExtendHeader = header
		req.Body = input
	}

	request, err := req.ToRequestContext(ctx)
	if err != nil {
		return mt.Zero[O](), nil, fmt.Errorf("toRequest: %w", err)
	}

	response, err := do.Do(request)
	if err != nil {
		return mt.Zero[O](), nil, NewBaseResponseError(err, req.Method, "")
	}
	if !req.Accepts(response.StatusCode) {
		return mt.Zero[O](), response, &BadStatusCodeError{Got: response.StatusCode}
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
