package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/adapter/ddnsmetric"
	"github.com/duakc/lightddns/adapter/ddnsprovider"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

// Aliyun Alidns API surface — RPC-style, signed with v1 (HMAC-SHA1).
// See sig_rpc.go for the signing logic and sig_roa.go for the ROA / V3
// scheme kept around for non-Alidns services.
//
// https://help.aliyun.com/zh/dns/api-alidns-2015-01-09-overview
const (
	AlidnsServiceName    = "alidns"
	AlidnsDefaultVersion = "2015-01-09"

	AlidnsActionDescribeDomains       = "DescribeDomains"
	AlidnsActionDescribeDomainRecords = "DescribeDomainRecords"
	AlidnsActionAddDomainRecord       = "AddDomainRecord"
	AlidnsActionUpdateDomainRecord    = "UpdateDomainRecord"
	AlidnsActionDeleteDomainRecord    = "DeleteDomainRecord"
)

const (
	opDescribeDomains = ddnsprovider.OpDescribeDomains
	opListRecords     = ddnsprovider.OpListRecords
	opCreateRecord    = ddnsprovider.OpCreateRecord
	opUpdateRecord    = ddnsprovider.OpUpdateRecord
	opDeleteRecord    = ddnsprovider.OpDeleteRecord
)

var AliyunDNSEndpoint = mt.Must(urlpkg.Parse("https://alidns.aliyuncs.com"))

var _ ddnsx.DomainSearcher = (*Client)(nil)

type Client struct {
	logger *zap.Logger
	do     httpx.HTTPRequester

	metricsRouter *ddnsprovider.ApiMetricsRouter
}

func NewClient(ctx context.Context, logger *zap.Logger,
	providerName string,
	do httpx.HTTPRequester,
	secretAccessKeyId, secretAccessKeySecret, secretSecurityToken string,
) *Client {
	router := ddnsprovider.NewMetricsRouter(
		services.Lookup[ddnsmetric.ProviderFactory](ctx), providerName)
	router.RegisterDefault()

	return &Client{
		logger:        logger,
		metricsRouter: router,
		// Alidns uses the RPC-style v1 signature: all public params (and
		// the resulting Signature) ride on the URL query string. The
		// ROA-style signer in [AliyunROASignClient] is kept in the package
		// for reference / future use against ROA services.
		do: &AliyunRPCSignClient{
			HTTPRequester:         do,
			Logger:                logger,
			SecretSecurityToken:   secretSecurityToken,
			SecretAccessKeyId:     secretAccessKeyId,
			SecretAccessKeySecret: secretAccessKeySecret,
		},
	}
}

// DescribeDomains — https://help.aliyun.com/document_detail/29751.html
func (c *Client) DescribeDomains(ctx context.Context,
	req DescribeDomainsRequest,
) (resp DescribeDomainsResponse, err error) {
	defer c.metricsRouter.RecordAPI(opDescribeDomains)(&err)
	return doAction[DescribeDomainsResponse](ctx, c, AlidnsActionDescribeDomains, req.Query())
}

// DescribeDomainRecords — https://help.aliyun.com/document_detail/29774.html
func (c *Client) DescribeDomainRecords(ctx context.Context,
	req DescribeDomainRecordsRequest,
) (resp DescribeDomainRecordsResponse, err error) {
	defer c.metricsRouter.RecordAPI(opListRecords)(&err)
	return doAction[DescribeDomainRecordsResponse](ctx, c, AlidnsActionDescribeDomainRecords, req.Query())
}

// AddDomainRecord — https://help.aliyun.com/document_detail/29772.html
func (c *Client) AddDomainRecord(ctx context.Context,
	req AddDomainRecordRequest,
) (resp AddDomainRecordResponse, err error) {
	defer c.metricsRouter.RecordAPI(opCreateRecord)(&err)
	return doAction[AddDomainRecordResponse](ctx, c, AlidnsActionAddDomainRecord, req.Query())
}

// UpdateDomainRecord — https://help.aliyun.com/document_detail/29773.html
func (c *Client) UpdateDomainRecord(ctx context.Context,
	req UpdateDomainRecordRequest,
) (resp UpdateDomainRecordResponse, err error) {
	defer c.metricsRouter.RecordAPI(opUpdateRecord)(&err)
	return doAction[UpdateDomainRecordResponse](ctx, c, AlidnsActionUpdateDomainRecord, req.Query())
}

// DeleteDomainRecord — https://help.aliyun.com/document_detail/29771.html
func (c *Client) DeleteDomainRecord(ctx context.Context,
	req DeleteDomainRecordRequest,
) (resp DeleteDomainRecordResponse, err error) {
	defer c.metricsRouter.RecordAPI(opDeleteRecord)(&err)
	return doAction[DeleteDomainRecordResponse](ctx, c, AlidnsActionDeleteDomainRecord, req.Query())
}

func (c *Client) SearchDomain(ctx context.Context, keyword string) map[string]string {

	const pageSize = 100

	logger := c.actionLogger(AlidnsActionDescribeDomains).
		With(zap.String("keyword", keyword))
	logger.Info("search domain id from upstream")

	result := make(map[string]string)
	for pageNumber := 1; ; pageNumber++ {
		page, err := c.DescribeDomains(ctx, DescribeDomainsRequest{
			KeyWord:    keyword,
			PageNumber: pageNumber,
			PageSize:   pageSize,
		})
		if err != nil {
			logger.Warn("list domains failed", zap.Error(err))
			return nil
		}
		for _, di := range page.Domains.Domain {
			if !domains.IsDomainName(di.DomainName) {
				continue
			}
			result[di.DomainName] = di.DomainName
		}
		if len(page.Domains.Domain) < pageSize ||
			int64(pageNumber)*pageSize >= page.TotalCount {
			return result
		}
	}
}

func (c *Client) actionLogger(action string) *zap.Logger {
	return c.logger.With(zap.String("action", action))
}

func doAction[Resp any](ctx context.Context, c *Client, action string, query urlpkg.Values) (Resp, error) {
	var zero Resp

	logger := c.actionLogger(action)

	req := newRequest(action)
	maps.Copy(req.Query, query)

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
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if jerr := json.NewDecoder(resp.Body).Decode(apiErr); jerr != nil {
			logger.Warn("decode error response failed",
				zap.Int("status", resp.StatusCode),
				zap.Error(jerr),
			)
			return zero, &httpx.BadStatusCodeError{Got: resp.StatusCode}
		}
		logger.Warn("api returned error",
			zap.String("error_code", apiErr.Code),
			zap.String("error_message", apiErr.Message),
			zap.String("request_id", apiErr.RequestID),
		)
		return zero, apiErr
	}

	var out Resp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		logger.Warn("decode response failed",
			zap.Error(err),
		)

		return zero, fmt.Errorf("decode %s: %w", action, err)
	}
	return out, nil
}

func newRequest(action string) httpx.ReqConfig {
	r := httpx.NewReqConfig(http.MethodPost, AliyunDNSEndpoint)
	r.Query.Set(ParamAction, action)
	r.Query.Set(ParamVersion, AlidnsDefaultVersion)
	return r
}
