package httpxx

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/duakc/mt"
)

type dummyJSONMarshaler struct {
	V any
}

func (d *dummyJSONMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.V)
}

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
		// ToRequestContext infers Content-Type: application/json from the
		// json.Marshaler interface, so no header mutation is needed here.
		if inputJM, ok := input.(json.Marshaler); ok {
			req.Body = inputJM
		} else {
			req.Body = &dummyJSONMarshaler{input}
		}
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
	if ct := response.Header.Get("Content-Type"); !isJSONContentType(ct) {
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

func isJSONContentType(ct string) bool {
	if ct == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return false
	}
	return mediaType == "application/json"
}
