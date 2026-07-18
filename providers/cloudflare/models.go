package cloudflare

import (
	"errors"
	"fmt"

	"github.com/duakc/mt"
)

type BaseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BaseError) Error() string {
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

type MessageError struct {
	BaseError

	DocumentationUrl string `json:"documentation_url"`
}

type ResultInfo struct {
	Count      int `json:"count"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// Response is the common Cloudflare API response envelope.
type Response struct {
	Success  bool           `json:"success"`
	Errors   []MessageError `json:"errors"`
	Messages []MessageError `json:"messages"`
}

func (r Response) Error() error {
	return errors.Join(mt.Map(r.Errors, func(s MessageError) error {
		return &s
	})...)
}

func (r Response) JoinError(err error) error {
	return errors.Join(err, r.Error())
}

// Zone is one zone returned by ListZones.
type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Record is one DNS record returned by ListDNSRecords.
type Record struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type DNSRecordRequest struct {
	Name           string `json:"name"`
	Ttl            uint32 `json:"ttl"`
	Type           string `json:"type"`
	Comment        string `json:"comment"`
	Content        string `json:"content"`
	PrivateRouting bool   `json:"private_routing"`
	Proxied        bool   `json:"proxied"`
}

type ListZonesRequest struct {
	Status  string
	Name    string
	Page    int
	PerPage int
}

type ListZonesResponse struct {
	Response

	Result     []Zone     `json:"result"`
	ResultInfo ResultInfo `json:"result_info"`
}

type ListDNSRecordsRequest struct {
	ZoneID string
	Name   string
	Type   string

	Page    int
	PerPage int
}

type ListDNSRecordsResponse struct {
	Response

	Result     []Record   `json:"result"`
	ResultInfo ResultInfo `json:"result_info"`
}

type CreateDNSRecordRequest struct {
	ZoneID string
	Body   DNSRecordRequest
}

type CreateDNSRecordResponse struct {
	Response

	Result Record `json:"result"`
}

type UpdateDNSRecordRequest struct {
	ZoneID   string
	RecordID string
	Body     DNSRecordRequest
}

type UpdateDNSRecordResponse struct {
	Response

	Result Record `json:"result"`
}

type DeleteDNSRecordRequest struct {
	ZoneID   string
	RecordID string
}

type DeleteDNSRecordResult struct {
	ID string `json:"id"`
}

type DeleteDNSRecordResponse struct {
	Response

	Result DeleteDNSRecordResult `json:"result"`
}
