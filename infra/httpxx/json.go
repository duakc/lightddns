package httpxx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/common"
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
		return common.Zero[O](), nil, fmt.Errorf("toRequest: %w", err)
	}
	var cancel context.CancelFunc
	ctx, cancel = common.ContextWithTimeout(ctx, constpkg.DefaultHTTPTimeout)
	defer cancel()
	response, err := do.Do(request)
	if err != nil {
		return common.Zero[O](), nil, NewResponseError(req.Method, request.URL.String(), err)
	}
	if response.StatusCode != http.StatusOK {
		return common.Zero[O](), response, &BadStatusCodeError{Got: response.StatusCode}
	}
	if ct := response.Header.Get("Content-Type"); ct != "application/json" {
		return common.Zero[O](), response, fmt.Errorf("not a JSON response: Content-Type: %s", ct)
	}
	decoder := json.NewDecoder(response.Body)
	var output O
	err = decoder.Decode(&output)
	if err != nil {
		return common.Zero[O](), response, fmt.Errorf("decode JSON: %w", err)
	}
	return output, response, nil
}
