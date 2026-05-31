package internal

import (
	"fmt"
	"net/http"

	"github.com/duakc/mt/common/validator"
)

// APIError is the Error envelope Tencent Cloud returns inside the Response
// body when an action fails. It uses HTTP 200 even on error, so this is the
// authoritative failure signal.
type APIError struct {
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	RequestID string `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("tencentcloud api error: code=%s message=%s request_id=%s",
		e.Code, e.Message, e.RequestID)
}

type Common struct {
	// https://cloud.tencent.com/document/api/1427/56188
	// Action string
	// generate by call Headers
	// Timestamp string
	// Version string

	// optional
	Token     string
	Language  string
	Region    string
	Timestamp int64

	Host string
}

func (c *Common) Headers() (http.Header, error) {
	if err := validator.GreaterThan(c.Timestamp, 0, "Timestamp"); err != nil {
		return nil, err
	}

	h := make(http.Header)
	host := c.Host
	if host == "" {
		host = tencentCloudEndpoint.Host
	}
	h.Set("Host", host)

	h.Set("X-TC-Timestamp", fmt.Sprintf("%d", c.Timestamp))

	if len(c.Token) > 0 {
		h.Set("X-TC-Token", c.Token)
	}
	if len(c.Language) > 0 {
		h.Set("X-TC-Language", c.Language)
	}
	if len(c.Region) > 0 {
		h.Set("X-TC-Region", c.Region)
	}

	return h, nil
}

// Record is one DNS record returned by DescribeRecordList.
//
// https://cloud.tencent.com/document/api/1427/56166
type Record struct {
	RecordId      uint64 `json:"RecordId"`
	Name          string `json:"Name"`   // subdomain (host part), e.g. "www"; "@" means apex
	Type          string `json:"Type"`   // A, AAAA, CNAME...
	Line          string `json:"Line"`   // line name (Chinese), e.g. "默认"
	LineId        string `json:"LineId"` // line id, e.g. "0"
	Value         string `json:"Value"`  // record content
	TTL           uint32 `json:"TTL"`
	MX            uint32 `json:"MX"`
	Status        string `json:"Status"`
	Weight        *int   `json:"Weight,omitempty"`
	MonitorStatus string `json:"MonitorStatus"`
	Remark        string `json:"Remark"`
	UpdatedOn     string `json:"UpdatedOn"`
	DomainId      uint64 `json:"DomainId"`
}

// DefaultRecordLine is the line used for DDNS records: 默认 (default).
//
// Tencent's API expects the line name in Chinese; this is the documented
// value for the catch-all line.
const DefaultRecordLine = "默认"

// CreateRecordRequest is the request body for CreateRecord.
//
// https://cloud.tencent.com/document/api/1427/56180
type CreateRecordRequest struct {
	Domain     string `json:"Domain"`
	SubDomain  string `json:"SubDomain"`
	RecordType string `json:"RecordType"`
	RecordLine string `json:"RecordLine"`
	Value      string `json:"Value"`
	TTL        uint32 `json:"TTL,omitempty"`
}

// ModifyRecordRequest is the request body for ModifyRecord.
//
// https://cloud.tencent.com/document/api/1427/56157
type ModifyRecordRequest struct {
	Domain     string `json:"Domain"`
	RecordId   uint64 `json:"RecordId"`
	SubDomain  string `json:"SubDomain"`
	RecordType string `json:"RecordType"`
	RecordLine string `json:"RecordLine"`
	Value      string `json:"Value"`
	TTL        uint32 `json:"TTL,omitempty"`
}

type DomainInfo struct {
	CNAMESpeedup     string   `json:"CNAMESpeedup"`
	CreatedOn        string   `json:"CreatedOn"`
	DNSStatus        string   `json:"DNSStatus"`
	DomainId         int      `json:"DomainId"`
	EffectiveDNS     []string `json:"EffectiveDNS"`
	Grade            string   `json:"Grade"`
	GradeLevel       int      `json:"GradeLevel"`
	GradeTitle       string   `json:"GradeTitle"`
	GroupId          int      `json:"GroupId"`
	IsVip            string   `json:"IsVip"`
	Name             string   `json:"Name"`
	Owner            string   `json:"Owner"`
	Punycode         string   `json:"Punycode"`
	RecordCount      int      `json:"RecordCount"`
	Remark           string   `json:"Remark"`
	SearchEnginePush string   `json:"SearchEnginePush"`
	Status           string   `json:"Status"`
	TTL              int      `json:"TTL"`
	TagList          []string `json:"TagList"`
	UpdatedOn        string   `json:"UpdatedOn"`
	VipAutoRenew     string   `json:"VipAutoRenew"`
	VipEndAt         string   `json:"VipEndAt"`
	VipStartAt       string   `json:"VipStartAt"`
}
