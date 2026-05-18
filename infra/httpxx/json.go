package httpxx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/duakc/mt"
)

type dummyJSONMarshaler struct {
	V any
}

func (d *dummyJSONMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.V)
}

func JSONRequest[O any](ctx context.Context, do HTTPRequester, req ReqConfig, input any) (O, *http.Response, error) {
	if input != nil {
		req.ExtendHeader.Set("Content-Type", "application/json")
		var JM json.Marshaler
		if inputJM, ok := input.(json.Marshaler); ok {
			JM = inputJM
		} else {
			JM = &dummyJSONMarshaler{input}
		}
		req.Body = JM
	}

	request, err := req.ToRequestContext(ctx)
	if err != nil {
		return mt.Zero[O](), nil, fmt.Errorf("toRequest: %w", err)
	}

	response, err := do.Do(request)
	if err != nil {
		return mt.Zero[O](), nil, NewBaseResponseError(err, req.Method, "")
	}
	if response.StatusCode != http.StatusOK {
		return mt.Zero[O](), response, &BadStatusCodeError{Got: response.StatusCode}
	}
	if ct := response.Header.Get("Content-Type"); ct != "application/json" {
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
