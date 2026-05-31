package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"
	"github.com/duakc/lightddns/infra/zaplog"

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

var tencentCloudEndpoint = mt.Must(urlpkg.Parse("https://dnspod.tencentcloudapi.com"))

type Client struct {
	do httpx.HTTPRequester
}

func NewClient(do httpx.HTTPRequester, secretId, secretKey string) *Client {
	return &Client{
		do: &TencentSignClient{
			HTTPRequester: do,
			SecretId:      secretId,
			SecretKey:     secretKey,
			Service:       DNSPodServiceName,
		},
	}
}

func (c *Client) newRequest(method, action string) httpx.ReqConfig {
	req := httpx.NewReqConfig(method, tencentCloudEndpoint)
	req.ExtendHeader.Set("X-TC-Action", action)
	req.ExtendHeader.Set("X-TC-Version", DNSPodDefaultVersion)
	return req
}

func jsonAction[T any](ctx context.Context, c *Client, action string, body map[string]any) (T, error) {
	var zero T
	logger := zaplog.FromOrPackage(ctx, "tencentcloud", "internal").
		With(zap.String("action", action))
	logger.Debug("tencentcloud: api call start")

	req := c.newRequest(http.MethodPost, action)
	req.ExtendHeader.Set("Content-Type", ContentTypeJson)
	req.Body = body

	httpReq, err := req.ToRequestContext(ctx)
	if err != nil {
		return zero, fmt.Errorf("build request %s: %w", action, err)
	}
	resp, err := c.do.Do(httpReq)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		logger.Warn("tencentcloud: send request failed",
			zap.Error(err))
		return zero, fmt.Errorf("send %s: %w", action, err)
	}

	if resp == nil {
		panic("empty response with error")
	}

	if resp.StatusCode >= 400 {
		logger.Warn("tencentcloud: bad status code",
			zap.Int("status", resp.StatusCode))
		return zero, &httpx.BadStatusCodeError{Got: resp.StatusCode}
	}

	// Read the raw body so we can log it on decode failure or API error.
	// Bodies are small (a single Response envelope) so this is cheap.
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("read body %s: %w", action, err)
	}

	var out Response[T]
	if err := json.Unmarshal(rawBody, &out); err != nil {
		logger.Warn("tencentcloud: decode response failed",
			zap.ByteString("body", rawBody),
			zap.Error(err))
		return zero, fmt.Errorf("decode %s: %w", action, err)
	}
	if out.Error != nil {
		logger.Warn("tencentcloud: api returned error",
			zap.String("error_code", out.Error.Code),
			zap.String("error_message", out.Error.Message),
			zap.String("request_id", out.RequestID))
		return zero, out.Error
	}
	logger.Debug("tencentcloud: api call ok",
		zap.String("request_id", out.RequestID))
	return out.Data, nil
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

func (c *Client) DomainInfo(ctx context.Context, search string) (DomainInfo, error) {
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

	// tryKeyword: walk all pages for one keyword; short-circuit on exact match.
	tryKeyword := func(keyword string) (DomainInfo, bool, error) {
		for offset := 0; ; {
			page, err := fetchPage(keyword, offset)
			if err != nil {
				return DomainInfo{}, false, err
			}
			for _, di := range page.DomainList {
				if domains.IsSubDomain(search, di.Name) {
					return di, true, nil
				}
			}
			offset += len(page.DomainList)
			if len(page.DomainList) == 0 || offset >= page.DomainCountInfo.DomainTotal {
				return DomainInfo{}, false, nil
			}
		}
	}

	for _, keyword := range append(domains.CutFromHead(search), "") {
		di, found, err := tryKeyword(keyword)
		if err != nil {
			return DomainInfo{}, err
		}
		if found {
			return di, nil
		}
	}
	return DomainInfo{}, fmt.Errorf("domain not found: %s", search)
}
