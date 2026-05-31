//go:build goexperiment.jsonv2

// Build With: `GOEXPERIMENT=jsonv2 go build ...`
package internal

import (
	"encoding/json"
)

type Response[T any] struct {
	Data      T         `json:"inline"`
	Error     *APIError `json:"Error,omitempty"`
	RequestID string    `json:"RequestId"`
}

func (r *Response[T]) UnmarshalJSON(data []byte) error {
	type _Alias Response[T]
	return json.Unmarshal(data, (*_Alias[T])(r))
}

func (r *Response[T]) MarshalJSON() ([]byte, error) {
	type _Alias Response[T]
	return json.Marshal((*_Alias[T])(r))
}
