package internal

import (
	"errors"
	"fmt"

	"github.com/duakc/lightddns/infra/common"
)

type BaseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BaseError) Error() string {
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

type Message struct {
	BaseError
	DocumentationUrl string `json:"documentation_url"`
	// other doesn't needed fields are ignored
}

type ResultInfo struct {
	Count      int `json:"count"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

type Response struct {
	Success  bool      `json:"success"`
	Errors   []Message `json:"errors"`
	Messages []Message `json:"messages"`
}

func (r *Response) Error() error {
	return errors.Join(common.Map(r.Errors, func(s Message) error {
		return &s
	})...)
}

func (r *Response) JoinError(err error) error {
	E := r.Error()
	return errors.Join(err, E)
}

type ResponseWithResult[T any] struct {
	Response `json:",inline"`
	Result   T `json:"result"`
}

type ResponseWithPage[T any] struct {
	Response   `json:",inline"`
	Result     []T        `json:"result"`
	ResultInfo ResultInfo `json:"result_info"`
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// other doesn't needed fields are ignored
}

type DNSRecord struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	ID      string `json:"id"`
	TTL     int    `json:"ttl"`
}

type UpdateDNSRecordRequest struct {
	Name           string `json:"name"`
	Ttl            int    `json:"ttl"`
	Type           string `json:"type"`
	Comment        string `json:"comment"`
	Content        string `json:"content"`
	PrivateRouting bool   `json:"private_routing"`
	Proxied        bool   `json:"proxied"`
}
