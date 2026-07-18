package aliyun

import (
	"fmt"
	urlpkg "net/url"
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

// Domain is one zone returned by DescribeDomains.
type Domain struct {
	DomainName string `json:"DomainName"`
	DomainId   string `json:"DomainId"`

	// uncomment when needed.
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
//
// All fields are optional; unset (zero) fields are omitted so the server
// applies its defaults (PageNumber=1, PageSize=20, no keyword filter).
type DescribeDomainsRequest struct {
	KeyWord    string
	PageNumber int
	PageSize   int
}

func (r DescribeDomainsRequest) Query() urlpkg.Values {
	p := NewParams()
	if r.KeyWord != "" {
		Add(p, "KeyWord", r.KeyWord)
	}
	if r.PageNumber > 0 {
		Add(p, "PageNumber", r.PageNumber)
	}
	if r.PageSize > 0 {
		Add(p, "PageSize", r.PageSize)
	}
	return p.Query()
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
//
// DomainName is required. RRKeyWord / TypeKeyWord narrow the result set and
// are added only when set; PageNumber / PageSize fall back to server
// defaults (1 / 20) when zero.
type DescribeDomainRecordsRequest struct {
	DomainName  string
	RRKeyWord   string
	TypeKeyWord string
	PageNumber  int
	PageSize    int
}

func (r DescribeDomainRecordsRequest) Query() urlpkg.Values {
	p := NewParams()
	Add(p, "DomainName", r.DomainName)
	if r.RRKeyWord != "" {
		Add(p, "RRKeyWord", r.RRKeyWord)
	}
	if r.TypeKeyWord != "" {
		Add(p, "TypeKeyWord", r.TypeKeyWord)
	}
	if r.PageNumber > 0 {
		Add(p, "PageNumber", r.PageNumber)
	}
	if r.PageSize > 0 {
		Add(p, "PageSize", r.PageSize)
	}
	return p.Query()
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
//
// DomainName / RR / Type / Value are required. Line falls back to "default"
// on the server when omitted; TTL falls back to the zone's configured
// default (commonly 600s).
type AddDomainRecordRequest struct {
	DomainName string
	RR         string
	Type       string
	Value      string
	Line       string
	TTL        uint32
}

func (r AddDomainRecordRequest) Query() urlpkg.Values {
	p := NewParams()
	Add(p, "DomainName", r.DomainName)
	Add(p, "RR", r.RR)
	Add(p, "Type", r.Type)
	Add(p, "Value", r.Value)
	if r.Line != "" {
		Add(p, "Line", r.Line)
	}
	if r.TTL > 0 {
		Add(p, "TTL", r.TTL)
	}
	return p.Query()
}

type AddDomainRecordResponse struct {
	RecordId string `json:"RecordId"`
}

// UpdateDomainRecordRequest — https://help.aliyun.com/document_detail/29773.html
//
// RecordId / RR / Type / Value are required (the API treats the call as a
// full record replacement, not a patch). Line / TTL retain their existing
// values on the server when omitted.
type UpdateDomainRecordRequest struct {
	RecordId string
	RR       string
	Type     string
	Value    string
	Line     string
	TTL      uint32
}

func (r UpdateDomainRecordRequest) Query() urlpkg.Values {
	p := NewParams()
	Add(p, "RecordId", r.RecordId)
	Add(p, "RR", r.RR)
	Add(p, "Type", r.Type)
	Add(p, "Value", r.Value)
	if r.Line != "" {
		Add(p, "Line", r.Line)
	}
	if r.TTL > 0 {
		Add(p, "TTL", r.TTL)
	}
	return p.Query()
}

type UpdateDomainRecordResponse struct {
	RecordId string `json:"RecordId"`
}

// DeleteDomainRecordRequest — https://help.aliyun.com/document_detail/29771.html
type DeleteDomainRecordRequest struct {
	RecordId string
}

func (r DeleteDomainRecordRequest) Query() urlpkg.Values {
	p := NewParams()
	Add(p, "RecordId", r.RecordId)
	return p.Query()
}

type DeleteDomainRecordResponse struct {
	RecordId string `json:"RecordId"`
}
