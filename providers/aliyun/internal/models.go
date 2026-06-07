package internal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// DefaultRecordLine is the line used for DDNS records.
//
// Aliyun's DNS API expects an English keyword for the line; "default" is the
// catch-all line that resolves for all networks.
const DefaultRecordLine = "default"

// Record host name used for an apex (zone-root) record in Aliyun's API.
const ApexRecordHost = "@"

// APIError is the error envelope Aliyun returns when an action fails. The
// HTTP status code is 4xx / 5xx in this case and the JSON body carries Code
// / Message / RequestId. It is the authoritative failure signal.
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"Code"`
	Message    string `json:"Message"`
	RequestID  string `json:"RequestId"`
	HostID     string `json:"HostId"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("aliyun api error: status=%d code=%s message=%s request_id=%s",
		e.StatusCode, e.Code, e.Message, e.RequestID)
}

// Common carries values that go into the request headers for every Aliyun V3
// call. Zero values are filled lazily by Headers so callers can reuse the
// struct as a default.
type Common struct {
	SecretSecurityToken string

	Nonce string
	Time  time.Time
}

func (c Common) Headers() http.Header {
	if c.Time.IsZero() {
		c.Time = time.Now()
	}
	if c.Nonce == "" {
		c.Nonce = uuid.NewString()
	}

	header := make(http.Header)
	header.Set(HeaderDate, c.Time.UTC().Format(time.RFC3339))
	header.Set(HeaderSignatureNonce, c.Nonce)
	if c.SecretSecurityToken != "" {
		header.Set(HeaderSecurityToken, c.SecretSecurityToken)
	}
	return header
}

// Domain is one zone returned by DescribeDomains.
type Domain struct {
	DomainName string `json:"DomainName"`

	// uncomment when needed.
	// DomainId    string `json:"DomainId"`
	// AliDomain   bool   `json:"AliDomain"`
	// RecordCount int    `json:"RecordCount"`
	// CreateTime  string `json:"CreateTime"`
	// PunyCode    string `json:"PunyCode"`
	// VersionCode string `json:"VersionCode"`
	// VersionName string `json:"VersionName"`
	// InstanceId  string `json:"InstanceId"`
	// GroupId     string `json:"GroupId"`
	// GroupName   string `json:"GroupName"`
	// Remark      string `json:"Remark"`
	// Starmark    bool   `json:"Starmark"`
}

// Record is one DNS record returned by DescribeDomainRecords.
type Record struct {
	RecordId string `json:"RecordId"`
	RR       string `json:"RR"`
	Type     string `json:"Type"`
	Value    string `json:"Value"`
	Line     string `json:"Line"`

	// uncomment when needed.
	// DomainName string `json:"DomainName"`
	// TTL        uint32 `json:"TTL"`
	// Priority   uint32 `json:"Priority"`
	// Status     string `json:"Status"`
	// Locked     bool   `json:"Locked"`
	// Weight     int    `json:"Weight"`
	// Remark     string `json:"Remark"`
	// UpdateTime string `json:"UpdateTime"`
	// CreateTime string `json:"CreateTime"`
}

// DescribeDomainsRequest — https://help.aliyun.com/document_detail/29751.html
type DescribeDomainsRequest struct {
	KeyWord    string `json:"KeyWord,omitempty"`
	PageNumber int    `json:"PageNumber,omitempty"`
	PageSize   int    `json:"PageSize,omitempty"`
}

type DescribeDomainsResponse struct {
	TotalCount int64 `json:"TotalCount"`
	PageNumber int64 `json:"PageNumber"`
	PageSize   int64 `json:"PageSize"`
	Domains    struct {
		Domain []Domain `json:"Domain"`
	} `json:"Domains"`
}

// DescribeDomainRecordsRequest — https://help.aliyun.com/document_detail/29774.html
type DescribeDomainRecordsRequest struct {
	DomainName  string `json:"DomainName"`
	RRKeyWord   string `json:"RRKeyWord,omitempty"`
	TypeKeyWord string `json:"TypeKeyWord,omitempty"`
	PageNumber  int    `json:"PageNumber,omitempty"`
	PageSize    int    `json:"PageSize,omitempty"`
}

type DescribeDomainRecordsResponse struct {
	TotalCount    int64 `json:"TotalCount"`
	PageNumber    int64 `json:"PageNumber"`
	PageSize      int64 `json:"PageSize"`
	DomainRecords struct {
		Record []Record `json:"Record"`
	} `json:"DomainRecords"`
}

// AddDomainRecordRequest — https://help.aliyun.com/document_detail/29772.html
type AddDomainRecordRequest struct {
	DomainName string `json:"DomainName"`
	RR         string `json:"RR"`
	Type       string `json:"Type"`
	Value      string `json:"Value"`
	Line       string `json:"Line,omitempty"`
	TTL        uint32 `json:"TTL,omitempty"`
}

type AddDomainRecordResponse struct {
	RecordId string `json:"RecordId"`
}

// UpdateDomainRecordRequest — https://help.aliyun.com/document_detail/29773.html
type UpdateDomainRecordRequest struct {
	RecordId string `json:"RecordId"`
	RR       string `json:"RR"`
	Type     string `json:"Type"`
	Value    string `json:"Value"`
	Line     string `json:"Line,omitempty"`
	TTL      uint32 `json:"TTL,omitempty"`
}

type UpdateDomainRecordResponse struct {
	RecordId string `json:"RecordId"`
}

// DeleteDomainRecordRequest — https://help.aliyun.com/document_detail/29771.html
type DeleteDomainRecordRequest struct {
	RecordId string `json:"RecordId"`
}

type DeleteDomainRecordResponse struct {
	RecordId string `json:"RecordId"`
}
