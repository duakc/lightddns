//go:build goexperiment.jsonv2

// Build With: `GOEXPERIMENT=jsonv2 go build ...`
package internal

import _ "encoding/json"

type Response[T any] struct {
	Data      T      `json:"inline"`
	RequestID string `json:"RequestId"`
}
