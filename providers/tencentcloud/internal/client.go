package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/adapter/ddnsmetric"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

const (
	// ContentTypeUrlEncoded GET
	ContentTypeUrlEncoded = "application/x-www-form-urlencoded"

	// ContentTypeJson ContentTypeFormData POST
	ContentTypeJson     = "application/json; charset=utf-8"
	ContentTypeFormData = "multipart/form-data"
)

const (
	DNSPodServiceName = "dnspod"

	DNSPodActionDescribeDomainFilterList = "DescribeDomainFilterList"
	DNSPodActionDescribeRecordList       = "DescribeRecordList"
	DNSPodActionCreateRecord             = "CreateRecord"
	DNSPodActionModifyRecord             = "ModifyRecord"
	DNSPodActionDeleteRecord             = "DeleteRecord"

	DNSPodDefaultVersion = "2021-03-23"

	// DNSPodErrCodeNoDataOfRecord is returned by DescribeRecordList when the
	// domain exists but has no records matching the filter. For DDNS this is
	// the normal "first run" state, not an error — callers should treat it
	// as an empty list.
	DNSPodErrCodeNoDataOfRecord = "ResourceNotFound.NoDataOfRecord"
)

// Operation labels recorded against the provider API metrics vec. One label
// per HTTP request.
const (
	opDescribeDomains = "describe_domains"
	opListRecords     = "list_records"
	opCreateRecord    = "create_record"
	opModifyRecord    = "modify_record"
	opDeleteRecord    = "delete_record"
)

// metricOpByAction maps the Tencent DNSPod action name to the metric label
// recorded for that call.
var metricOpByAction = map[string]string{
	DNSPodActionDescribeDomainFilterList: opDescribeDomains,
	DNSPodActionDescribeRecordList:       opListRecords,
	DNSPodActionCreateRecord:             opCreateRecord,
	DNSPodActionModifyRecord:             opModifyRecord,
	DNSPodActionDeleteRecord:             opDeleteRecord,
}

const (
	HeaderAction  = "X-TC-Action"
	HeaderVersion = "X-TC-Version"

	HeaderTimestamp = "X-TC-Timestamp"
	HeaderToken     = "X-TC-Token"
	HeaderLanguage  = "X-TC-Language"
	HeaderRegion    = "X-TC-Region"
)

var TencentCloudEndpoint = mt.Must(urlpkg.Parse("https://dnspod.tencentcloudapi.com"))

var _ ddnsx.DomainIdFetcher = (*Client)(nil)

type Client struct {
	logger       *zap.Logger
	providerName string
	do           httpx.HTTPRequester

	metricsRouter *ddnsmetric.ProviderAPIRouter
}

func NewClient(logger *zap.Logger, providerName string,
	do httpx.HTTPRequester, secretId, secretKey string,
) *Client {
	return &Client{
		logger:       logger,
		providerName: providerName,
		do: &TencentSignClient{
			HTTPRequester: do,
			SecretId:      secretId,
			SecretKey:     secretKey,
			Service:       DNSPodServiceName,
		},
	}
}

// RegisterMetrics builds the per-op metric router. Must be called once during
// the owning provider's PreStart, before any API method fires — otherwise
// jsonAction's defer dereferences a nil router.
func (c *Client) RegisterMetrics(factory ddnsmetric.Factory) {
	c.metricsRouter = ddnsmetric.ProviderLeaf.NewRouter(factory, c.providerName, []string{
		opDescribeDomains,
		opListRecords,
		opCreateRecord,
		opModifyRecord,
		opDeleteRecord,
	})
}

func (c *Client) newRequest(method, action string) httpx.ReqConfig {
	req := httpx.NewReqConfig(method, TencentCloudEndpoint)
	req.ExtendHeader.Set(HeaderAction, action)
	req.ExtendHeader.Set(HeaderVersion, DNSPodDefaultVersion)
	return req
}

// DescribeRecordList https://cloud.tencent.com/document/api/1427/56166
func (c *Client) DescribeRecordList(ctx context.Context, domain, subDomain, recordType string) ([]Record, error) {
	type _describeRecordListResponse struct {
		RecordCountInfo struct {
			SubdomainCount int `json:"SubdomainCount"`
			ListCount      int `json:"ListCount"`
			TotalCount     int `json:"TotalCount"`
		} `json:"RecordCountInfo"`
		RecordList []Record `json:"RecordList"`
	}

	const (
		action   = DNSPodActionDescribeRecordList
		pageSize = 100
	)

	var all []Record
	for offset := 0; ; {
		body := map[string]any{
			"Domain": domain,
			"Limit":  pageSize,
			"Offset": offset,
		}
		if subDomain != "" {
			body["Subdomain"] = subDomain
		}
		if recordType != "" {
			body["RecordType"] = recordType
		}
		page, err := jsonAction[_describeRecordListResponse](ctx, c, action, body)
		if err != nil {
			// Tencent returns NoDataOfRecord when the domain has no matching
			// records. For initial DDNS setup that's the expected state, not
			// a failure — surface it as an empty list so the caller proceeds
			// to CreateRecord instead of bailing.
			if apiErr, ok := errors.AsType[*APIError](err); ok &&
				apiErr.Code == DNSPodErrCodeNoDataOfRecord {

				return all, nil
			}
			return nil, err
		}
		all = append(all, page.RecordList...)
		offset += len(page.RecordList)
		if len(page.RecordList) == 0 || offset >= page.RecordCountInfo.TotalCount {
			return all, nil
		}
	}
}

// CreateRecord creates one DNS record.
//
// https://cloud.tencent.com/document/api/1427/56180
func (c *Client) CreateRecord(ctx context.Context, req CreateRecordRequest) error {
	body := map[string]any{
		"Domain":     req.Domain,
		"SubDomain":  req.SubDomain,
		"RecordType": req.RecordType,
		"RecordLine": req.RecordLine,
		"Value":      req.Value,
	}
	if req.TTL > 0 {
		body["TTL"] = req.TTL
	}
	type _createRecordResponse struct {
		RecordId uint64 `json:"RecordId"`
	}
	_, err := jsonAction[_createRecordResponse](ctx, c, DNSPodActionCreateRecord, body)
	return err
}

// ModifyRecord updates one DNS record.
//
// https://cloud.tencent.com/document/api/1427/56157
func (c *Client) ModifyRecord(ctx context.Context, req ModifyRecordRequest) error {
	body := map[string]any{
		"Domain":     req.Domain,
		"RecordId":   req.RecordId,
		"SubDomain":  req.SubDomain,
		"RecordType": req.RecordType,
		"RecordLine": req.RecordLine,
		"Value":      req.Value,
	}
	if req.TTL > 0 {
		body["TTL"] = req.TTL
	}
	type _modifyRecordResponse struct {
		RecordId uint64 `json:"RecordId"`
	}
	_, err := jsonAction[_modifyRecordResponse](ctx, c, DNSPodActionModifyRecord, body)
	return err
}

// DeleteRecord deletes one DNS record by id.
//
// https://cloud.tencent.com/document/api/1427/56176
func (c *Client) DeleteRecord(ctx context.Context, domain string, recordId uint64) error {
	body := map[string]any{
		"Domain":   domain,
		"RecordId": recordId,
	}
	type _deleteRecordResponse struct{}
	_, err := jsonAction[_deleteRecordResponse](ctx, c, DNSPodActionDeleteRecord, body)
	return err
}

// FetchDomainId implements [ddnsx.DomainIdFetcher]. Tencent's DNSPod APIs key
// records by the parent zone Name rather than a numeric id, so the returned
// map stores Name -> Name. [ddnsx.DomainIdCache] selects the longest suffix
// match for the queried FQDN and remembers any other domains seen along the
// way for future lookups. Each underlying page request is recorded against
// the API metric via jsonAction — no top-level metric is emitted.
//
// To minimise API calls the search walks suffixes of `search` (longest-first)
// and short-circuits as soon as one keyword turns up a domain that's a parent
// of `search`. Returns nil on transport / API failure so the cache treats it
// as "no result".
func (c *Client) FetchDomainId(ctx context.Context, search string) map[string]string {
	if mt.Done(ctx) || len(search) == 0 {
		return nil
	}
	type _domainInfoResponse struct {
		DomainCountInfo struct {
			AllTotal      int `json:"AllTotal"`
			DomainTotal   int `json:"DomainTotal"`
			ErrorTotal    int `json:"ErrorTotal"`
			GroupTotal    int `json:"GroupTotal"`
			LockTotal     int `json:"LockTotal"`
			MineTotal     int `json:"MineTotal"`
			PauseTotal    int `json:"PauseTotal"`
			ShareOutTotal int `json:"ShareOutTotal"`
			ShareTotal    int `json:"ShareTotal"`
			SpamTotal     int `json:"SpamTotal"`
			VipExpire     int `json:"VipExpire"`
			VipTotal      int `json:"VipTotal"`
		} `json:"DomainCountInfo"`

		DomainList []DomainInfo `json:"DomainList"`
	}

	const (
		action   = DNSPodActionDescribeDomainFilterList
		pageSize = 100
	)

	logger := c.actionLogger(action).
		With(zap.String("search", search))

	logger.Info("search domain id from upstream")

	fetchPage := func(keyword string, offset int) (_domainInfoResponse, error) {
		body := map[string]any{
			"Type":   "ALL",
			"Limit":  pageSize,
			"Offset": offset,
		}
		if keyword != "" {
			body["Keyword"] = keyword
		}
		return jsonAction[_domainInfoResponse](ctx, c, action, body)
	}

	result := make(map[string]string)
	for _, keyword := range append(domains.CutFromHead(search), "") {
		matched := false
		for offset := 0; ; {
			page, err := fetchPage(keyword, offset)
			if err != nil {
				logger.Warn("tencentcloud: list domains failed",
					zap.String("keyword", keyword), zap.Error(err))
				return nil
			}
			for _, di := range page.DomainList {
				if !domains.IsDomainName(di.Name) {
					continue
				}
				result[di.Name] = di.Name
				if domains.IsSubDomain(search, di.Name) {
					matched = true
				}
			}
			offset += len(page.DomainList)
			if len(page.DomainList) == 0 || offset >= page.DomainCountInfo.DomainTotal {
				break
			}
		}
		if matched {
			return result
		}
	}
	return result
}

func (c *Client) actionLogger(action string) *zap.Logger {
	return c.logger.With(zap.String("action", action))
}

func jsonAction[T any](ctx context.Context, c *Client, action string, body map[string]any) (_ T, err error) {
	var zero T
	defer c.metricsRouter.RecordAPI(metricOpByAction[action])(&err)

	logger := c.actionLogger(action)
	logger.Debug("tencentcloud: api call start")

	req := c.newRequest(http.MethodPost, action)
	req.ExtendHeader.Set("Content-Type", ContentTypeJson)
	req.Body = body

	httpReq, perr := req.ToRequestContext(ctx)
	if perr != nil {
		err = fmt.Errorf("build request %s: %w", action, perr)
		return zero, err
	}
	resp, perr := c.do.Do(httpReq)
	if resp != nil {
		defer resp.Body.Close()
	}
	if perr != nil {
		logger.Warn("tencentcloud: send request failed",
			zap.Error(perr))
		err = fmt.Errorf("send %s: %w", action, perr)
		return zero, err
	}

	if resp == nil {
		panic("empty response with error")
	}

	if resp.StatusCode >= 400 {
		logger.Warn("tencentcloud: bad status code",
			zap.Int("status", resp.StatusCode))
		err = &httpx.BadStatusCodeError{Got: resp.StatusCode}
		return zero, err
	}

	// Read the raw body so we can log it on decode failure or API error.
	// Bodies are small (a single Response envelope) so this is cheap.
	rawBody, perr := io.ReadAll(resp.Body)
	if perr != nil {
		err = fmt.Errorf("read body %s: %w", action, perr)
		return zero, err
	}

	var out Response[T]
	if perr := json.Unmarshal(rawBody, &out); perr != nil {
		logger.Warn("tencentcloud: decode response failed",
			zap.ByteString("body", rawBody),
			zap.Error(perr))
		err = fmt.Errorf("decode %s: %w", action, perr)
		return zero, err
	}
	if out.Error != nil {
		logger.Warn("tencentcloud: api returned error",
			zap.String("error_code", out.Error.Code),
			zap.String("error_message", out.Error.Message),
			zap.String("request_id", out.RequestID))
		err = out.Error
		return zero, err
	}
	logger.Debug("tencentcloud: api call ok",
		zap.String("request_id", out.RequestID))
	return out.Data, nil
}
