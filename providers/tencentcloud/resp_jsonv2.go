//go:build goexperiment.jsonv2

// Build With: `GOEXPERIMENT=jsonv2 go build ...`
package tencentcloud

import (
	"encoding/json"
	"fmt"
)

type Response[T any] struct {
	Data      T
	Error     *APIError
	RequestID string `json:"RequestId"`
}

func (r *Response[T]) UnmarshalJSON(data []byte) error {
	var outer struct {
		Resp json.RawMessage `json:"Response"`
	}
	if err := json.Unmarshal(data, &outer); err != nil {
		return fmt.Errorf("unmarshal outer Response: %w", err)
	}

	var meta struct {
		RequestID string    `json:"RequestId"`
		Error     *APIError `json:"Error"`
	}
	if err := json.Unmarshal(outer.Resp, &meta); err != nil {
		return fmt.Errorf("unmarshal RequestId: %w", err)
	}

	r.RequestID = meta.RequestID
	if meta.Error != nil {
		meta.Error.RequestID = meta.RequestID
		r.Error = meta.Error
		return nil
	}
	if err := json.Unmarshal(outer.Resp, &r.Data); err != nil {
		return fmt.Errorf("unmarshal Data: %w", err)
	}
	return nil
}

func (r *Response[T]) MarshalJSON() ([]byte, error) {
	dataBytes, err := json.Marshal(r.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal Data: %w", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(dataBytes, &fields); err != nil {
		return nil, fmt.Errorf("unmarshal Data into map: %w", err)
	}
	if fields == nil {
		fields = make(map[string]json.RawMessage)
	}

	reqIDBytes, err := json.Marshal(r.RequestID)
	if err != nil {
		return nil, err
	}
	fields["RequestId"] = reqIDBytes

	innerBytes, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	outer := map[string]json.RawMessage{"Response": innerBytes}
	return json.Marshal(outer)
}
