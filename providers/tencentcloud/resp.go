package tencentcloud

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"fmt"
)

type Response[T any] struct {
	Data      T
	Error     *APIError
	RequestID string `json:"RequestId"`
}

func (r *Response[T]) UnmarshalJSON(data []byte) error {
	var outer struct {
		Resp jsontext.Value `json:"Response"`
	}
	if err := jsonv2.Unmarshal(data, &outer); err != nil {
		return fmt.Errorf("unmarshal outer Response: %w", err)
	}

	var meta struct {
		RequestID string    `json:"RequestId"`
		Error     *APIError `json:"Error"`
	}
	if err := jsonv2.Unmarshal(outer.Resp, &meta); err != nil {
		return fmt.Errorf("unmarshal RequestId: %w", err)
	}

	r.RequestID = meta.RequestID
	if meta.Error != nil {
		meta.Error.RequestID = meta.RequestID
		r.Error = meta.Error

		//nolint:nilerr
		return nil
	}

	if err := jsonv2.Unmarshal(outer.Resp, &r.Data); err != nil {
		return fmt.Errorf("unmarshal Data: %w", err)
	}
	return nil
}

func (r *Response[T]) MarshalJSON() ([]byte, error) {
	dataBytes, err := jsonv2.Marshal(r.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal Data: %w", err)
	}

	var fields map[string]jsontext.Value
	if err := jsonv2.Unmarshal(dataBytes, &fields); err != nil {
		return nil, fmt.Errorf("unmarshal Data into map: %w", err)
	}
	if fields == nil {
		fields = make(map[string]jsontext.Value)
	}

	reqIDBytes, err := jsonv2.Marshal(r.RequestID)
	if err != nil {
		return nil, err
	}
	fields["RequestId"] = jsontext.Value(reqIDBytes)

	innerBytes, err := jsonv2.Marshal(fields)
	if err != nil {
		return nil, err
	}
	outer := map[string]jsontext.Value{"Response": jsontext.Value(innerBytes)}
	return jsonv2.Marshal(outer)
}
