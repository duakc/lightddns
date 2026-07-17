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

type Response struct {
	Success  bool           `json:"success"`
	Errors   []MessageError `json:"errors"`
	Messages []MessageError `json:"messages"`
}

func (r *Response) Error() error {
	return errors.Join(mt.Map(r.Errors, func(s MessageError) error {
		return &s
	})...)
}

func (r *Response) JoinError(err error) error {
	E := r.Error()
	return errors.Join(err, E)
}

type ResponseWithResult[T any] struct {
	Response `json:",inline"`

	Result T `json:"result"`
}

type ResponseWithPage[T any] struct {
	Response `json:",inline"`

	Result     []T        `json:"result"`
	ResultInfo ResultInfo `json:"result_info"`
}

// Zone is one zone returned by ListZones.
type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// uncomment when needed.
	// Status              string   `json:"status"`
	// Paused              bool     `json:"paused"`
	// Type                string   `json:"type"`
	// NameServers         []string `json:"name_servers"`
	// OriginalNameServers []string `json:"original_name_servers"`
	// CreatedOn           string   `json:"created_on"`
	// ModifiedOn          string   `json:"modified_on"`
	// ActivatedOn         string   `json:"activated_on"`
}

// DNSRecord is one DNS record returned by ListDNSRecords.
type DNSRecord struct {
	ID      string `json:"id"`
	Content string `json:"content"`

	// uncomment when needed.
	// Name       string `json:"name"`
	// Type       string `json:"type"`
	// TTL        int    `json:"ttl"`
	// Proxied    bool   `json:"proxied"`
	// Proxiable  bool   `json:"proxiable"`
	// ZoneID     string `json:"zone_id"`
	// ZoneName   string `json:"zone_name"`
	// Comment    string `json:"comment"`
	// CreatedOn  string `json:"created_on"`
	// ModifiedOn string `json:"modified_on"`
}

type UpdateDNSRecordRequest struct {
	Name           string `json:"name"`
	Ttl            uint32 `json:"ttl"`
	Type           string `json:"type"`
	Comment        string `json:"comment"`
	Content        string `json:"content"`
	PrivateRouting bool   `json:"private_routing"`
	Proxied        bool   `json:"proxied"`
}
