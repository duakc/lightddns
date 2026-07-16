package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/adapter/metricx"
	"github.com/duakc/lightddns/adapter/providerx"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

// Tencent Cloud DNSPod API surface — we only use the v3 API.
//
// https://cloud.tencent.com/document/api/1427/56188
const (
	DNSPodServiceName    = "dnspod"
	DNSPodDefaultVersion = "2021-03-23"

	DNSPodActionDescribeDomainFilterList = "DescribeDomainFilterList"
	DNSPodActionDescribeRecordList       = "DescribeRecordList"
	DNSPodActionCreateRecord             = "CreateRecord"
	DNSPodActionModifyRecord             = "ModifyRecord"
	DNSPodActionDeleteRecord             = "DeleteRecord"

	// DNSPodErrCodeNoDataOfRecord is returned by DescribeRecordList when the
	// domain exists but has no records matching the filter. For DDNS this is
	// the normal "first run" state, not an error — callers should treat it
	// as an empty list.
	DNSPodErrCodeNoDataOfRecord = "ResourceNotFound.NoDataOfRecord"
)

const (
	HeaderAction    = "X-TC-Action"
	HeaderVersion   = "X-TC-Version"
	HeaderTimestamp = "X-TC-Timestamp"
	HeaderToken     = "X-TC-Token"
	HeaderLanguage  = "X-TC-Language"
	HeaderRegion    = "X-TC-Region"
)

const ContentTypeJSON = "application/json; charset=utf-8"

const (
	opDescribeDomains = providerx.OpDescribeDomains
	opListRecords     = providerx.OpListRecords
	opCreateRecord    = providerx.OpCreateRecord
	opModifyRecord    = providerx.OpUpdateRecord
	opDeleteRecord    = providerx.OpDeleteRecord
)

var TencentCloudEndpoint = mt.Must(urlpkg.Parse("https://dnspod.tencentcloudapi.com"))

var _ ddnsx.DomainSearcher = (*Client)(nil)

type Client struct {
	logger *zap.Logger
	do     httpx.HTTPRequester

	metricsRouter *providerx.ApiMetricsRouter
}

func NewClient(ctx context.Context, logger *zap.Logger,
	providerName string,
	do httpx.HTTPRequester, secretId, secretKey string,
) *Client {
	router := providerx.NewMetricsRouter(
		services.Lookup[metricx.ProviderFactory](ctx), providerName)
	router.RegisterDefault()

	return &Client{
		logger:        logger,
		metricsRouter: router,
		do: &TencentSignHTTPRequester{
			HTTPRequester: do,
			Logger:        logger,
			SecretId:      secretId,
			SecretKey:     secretKey,
			Service:       DNSPodServiceName,
		},
	}
}

// DescribeDomainFilterList — https://cloud.tencent.com/document/api/1427/56173
func (c *Client) DescribeDomainFilterList(ctx context.Context,
	req DescribeDomainFilterListRequest,
) (resp DescribeDomainFilterListResponse, err error) {
	defer c.metricsRouter.RecordAPI(opDescribeDomains)(&err)
	return doAction[DescribeDomainFilterListResponse](ctx, c, DNSPodActionDescribeDomainFilterList, req)
}

// DescribeRecordList — https://cloud.tencent.com/document/api/1427/56166
func (c *Client) DescribeRecordList(ctx context.Context,
	req DescribeRecordListRequest,
) (resp DescribeRecordListResponse, err error) {
	defer c.metricsRouter.RecordAPI(opListRecords)(&err)
	return doAction[DescribeRecordListResponse](ctx, c, DNSPodActionDescribeRecordList, req)
}

// CreateRecord — https://cloud.tencent.com/document/api/1427/56180
func (c *Client) CreateRecord(ctx context.Context,
	req CreateRecordRequest,
) (resp CreateRecordResponse, err error) {
	defer c.metricsRouter.RecordAPI(opCreateRecord)(&err)
	return doAction[CreateRecordResponse](ctx, c, DNSPodActionCreateRecord, req)
}

// ModifyRecord — https://cloud.tencent.com/document/api/1427/56157
func (c *Client) ModifyRecord(ctx context.Context,
	req ModifyRecordRequest,
) (resp ModifyRecordResponse, err error) {
	defer c.metricsRouter.RecordAPI(opModifyRecord)(&err)
	return doAction[ModifyRecordResponse](ctx, c, DNSPodActionModifyRecord, req)
}

// DeleteRecord — https://cloud.tencent.com/document/api/1427/56176
func (c *Client) DeleteRecord(ctx context.Context,
	req DeleteRecordRequest,
) (resp DeleteRecordResponse, err error) {
	defer c.metricsRouter.RecordAPI(opDeleteRecord)(&err)
	return doAction[DeleteRecordResponse](ctx, c, DNSPodActionDeleteRecord, req)
}

func (c *Client) SearchDomain(ctx context.Context, keyword string) map[string]string {
	const pageSize = 100

	logger := c.actionLogger(DNSPodActionDescribeDomainFilterList).
		With(zap.String("keyword", keyword))
	logger.Info("search domain id from upstream")

	result := make(map[string]string)
	for offset := 0; ; {
		page, err := c.DescribeDomainFilterList(ctx, DescribeDomainFilterListRequest{
			Type:    "ALL",
			Limit:   pageSize,
			Offset:  offset,
			Keyword: keyword,
		})
		if err != nil {
			logger.Warn("list domains failed", zap.Error(err))
			return nil
		}
		for _, di := range page.DomainList {
			if !domains.IsDomainName(di.Name) {
				continue
			}
			result[di.Name] = di.Name
		}
		offset += len(page.DomainList)
		if len(page.DomainList) == 0 || offset >= page.DomainCountInfo.DomainTotal {
			return result
		}
	}
}

func (c *Client) actionLogger(action string) *zap.Logger {
	return c.logger.With(zap.String("action", action))
}

func doAction[Resp any](ctx context.Context, c *Client, action string, body any) (Resp, error) {
	var zero Resp

	logger := c.actionLogger(action)

	req := newRequest(http.MethodPost, action)
	req.ExtendHeader.Set("Host", TencentCloudEndpoint.Host)
	req.ExtendHeader.Set("Content-Type", ContentTypeJSON)
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
		return zero, fmt.Errorf("send %s: %w", action, err)
	}
	if resp == nil {
		panic("empty response with no error")
	}
	if resp.StatusCode >= 400 {
		return zero, &httpx.BadStatusCodeError{Got: resp.StatusCode}
	}

	var out Response[Resp]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		logger.Warn("decode response failed", zap.Error(err))
		return zero, fmt.Errorf("decode %s: %w", action, err)
	}

	if out.Error != nil {
		logger.Warn("api returned error",
			zap.String("error_code", out.Error.Code),
			zap.String("error_message", out.Error.Message),
			zap.String("request_id", out.RequestID))
		return zero, out.Error
	}
	return out.Data, nil
}

func newRequest(method, action string) httpx.ReqConfig {
	req := httpx.NewReqConfig(method, TencentCloudEndpoint)
	req.ExtendHeader.Set(HeaderAction, action)
	req.ExtendHeader.Set(HeaderVersion, DNSPodDefaultVersion)
	return req
}
