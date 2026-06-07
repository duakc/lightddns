package internal

import (
	"fmt"
	"net/http"
	"time"
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
}

func (c *Common) Headers() http.Header {
	if c.Timestamp == 0 {
		c.Timestamp = time.Now().UTC().Unix()
	}

	h := make(http.Header)

	h.Set(HeaderTimestamp, fmt.Sprintf("%d", c.Timestamp))

	if len(c.Token) > 0 {
		h.Set(HeaderToken, c.Token)
	}
	if len(c.Language) > 0 {
		h.Set(HeaderLanguage, c.Language)
	}
	if len(c.Region) > 0 {
		h.Set(HeaderRegion, c.Region)
	}

	return h
}

// DefaultRecordLine is the line used for DDNS records: 默认 (default).
//
// Tencent's API expects the line name in Chinese; this is the documented
// value for the catch-all line.
const DefaultRecordLine = "默认"

// Record is one DNS record returned by DescribeRecordList.
//
// https://cloud.tencent.com/document/api/1427/56166
type Record struct {
	RecordId uint64 `json:"RecordId"`
	Type     string `json:"Type"`
	Line     string `json:"Line"`
	Value    string `json:"Value"`

	// uncomment when needed.
	// Name          string `json:"Name"`          // "@" means apex
	// LineId        string `json:"LineId"`
	// TTL           uint32 `json:"TTL"`
	// MX            uint32 `json:"MX"`
	// Status        string `json:"Status"`
	// Weight        *int   `json:"Weight,omitempty"`
	// MonitorStatus string `json:"MonitorStatus"`
	// Remark        string `json:"Remark"`
	// UpdatedOn     string `json:"UpdatedOn"`
	// DomainId      uint64 `json:"DomainId"`
}

// DomainInfo is one domain returned by DescribeDomainFilterList.
type DomainInfo struct {
	Name string `json:"Name"`

	// uncomment when needed.
	// CNAMESpeedup     string   `json:"CNAMESpeedup"`
	// CreatedOn        string   `json:"CreatedOn"`
	// DNSStatus        string   `json:"DNSStatus"`
	// DomainId         int      `json:"DomainId"`
	// EffectiveDNS     []string `json:"EffectiveDNS"`
	// Grade            string   `json:"Grade"`
	// GradeLevel       int      `json:"GradeLevel"`
	// GradeTitle       string   `json:"GradeTitle"`
	// GroupId          int      `json:"GroupId"`
	// IsVip            string   `json:"IsVip"`
	// Owner            string   `json:"Owner"`
	// Punycode         string   `json:"Punycode"`
	// RecordCount      int      `json:"RecordCount"`
	// Remark           string   `json:"Remark"`
	// SearchEnginePush string   `json:"SearchEnginePush"`
	// Status           string   `json:"Status"`
	// TTL              int      `json:"TTL"`
	// TagList          []string `json:"TagList"`
	// UpdatedOn        string   `json:"UpdatedOn"`
	// VipAutoRenew     string   `json:"VipAutoRenew"`
	// VipEndAt         string   `json:"VipEndAt"`
	// VipStartAt       string   `json:"VipStartAt"`
}

// DescribeDomainFilterListRequest — https://cloud.tencent.com/document/api/1427/56173
type DescribeDomainFilterListRequest struct {
	Type    string `json:"Type"`
	Limit   int    `json:"Limit,omitempty"`
	Offset  int    `json:"Offset,omitempty"`
	Keyword string `json:"Keyword,omitempty"`
}

type DescribeDomainFilterListResponse struct {
	DomainCountInfo struct {
		DomainTotal int `json:"DomainTotal"`

		// uncomment when needed.
		// AllTotal      int `json:"AllTotal"`
		// ErrorTotal    int `json:"ErrorTotal"`
		// GroupTotal    int `json:"GroupTotal"`
		// LockTotal     int `json:"LockTotal"`
		// MineTotal     int `json:"MineTotal"`
		// PauseTotal    int `json:"PauseTotal"`
		// ShareOutTotal int `json:"ShareOutTotal"`
		// ShareTotal    int `json:"ShareTotal"`
		// SpamTotal     int `json:"SpamTotal"`
		// VipExpire     int `json:"VipExpire"`
		// VipTotal      int `json:"VipTotal"`
	} `json:"DomainCountInfo"`
	DomainList []DomainInfo `json:"DomainList"`
}

// DescribeRecordListRequest — https://cloud.tencent.com/document/api/1427/56166
type DescribeRecordListRequest struct {
	Domain     string `json:"Domain"`
	Subdomain  string `json:"Subdomain,omitempty"`
	RecordType string `json:"RecordType,omitempty"`
	Limit      int    `json:"Limit,omitempty"`
	Offset     int    `json:"Offset,omitempty"`
}

type DescribeRecordListResponse struct {
	RecordCountInfo struct {
		TotalCount int `json:"TotalCount"`

		// uncomment when needed.
		// SubdomainCount int `json:"SubdomainCount"`
		// ListCount      int `json:"ListCount"`
	} `json:"RecordCountInfo"`
	RecordList []Record `json:"RecordList"`
}

// CreateRecordRequest — https://cloud.tencent.com/document/api/1427/56180
type CreateRecordRequest struct {
	Domain     string `json:"Domain"`
	SubDomain  string `json:"SubDomain"`
	RecordType string `json:"RecordType"`
	RecordLine string `json:"RecordLine"`
	Value      string `json:"Value"`
	TTL        uint32 `json:"TTL,omitempty"`
}

type CreateRecordResponse struct {
	RecordId uint64 `json:"RecordId"`
}

// ModifyRecordRequest — https://cloud.tencent.com/document/api/1427/56157
type ModifyRecordRequest struct {
	Domain     string `json:"Domain"`
	RecordId   uint64 `json:"RecordId"`
	SubDomain  string `json:"SubDomain"`
	RecordType string `json:"RecordType"`
	RecordLine string `json:"RecordLine"`
	Value      string `json:"Value"`
	TTL        uint32 `json:"TTL,omitempty"`
}

type ModifyRecordResponse struct {
	RecordId uint64 `json:"RecordId"`
}

// DeleteRecordRequest — https://cloud.tencent.com/document/api/1427/56176
type DeleteRecordRequest struct {
	Domain   string `json:"Domain"`
	RecordId uint64 `json:"RecordId"`
}

type DeleteRecordResponse struct{}
