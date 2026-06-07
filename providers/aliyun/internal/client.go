package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/adapter/ddnsmetric"
	"github.com/duakc/lightddns/adapter/ddnsprovider"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt"
	"github.com/duakc/mt/freebuf"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

// Aliyun Alidns API surface — V3 signature, RPC-style.
//
// https://help.aliyun.com/zh/dns/api-alidns-2015-01-09-overview
const (
	AlidnsServiceName    = "alidns"
	AlidnsDefaultVersion = "2015-01-09"

	AlidnsActionDescribeDomains       = AlidnsServiceName + ":" + "DescribeDomains"
	AlidnsActionDescribeDomainRecords = AlidnsServiceName + ":" + "DescribeDomainRecords"
	AlidnsActionAddDomainRecord       = AlidnsServiceName + ":" + "AddDomainRecord"
	AlidnsActionUpdateDomainRecord    = AlidnsServiceName + ":" + "UpdateDomainRecord"
	AlidnsActionDeleteDomainRecord    = AlidnsServiceName + ":" + "DeleteDomainRecord"

	// AlidnsErrCodeDomainRecordDuplicate is returned by UpdateDomainRecord
	// when the new value equals the current value. For DDNS reconcile loops
	// that race against themselves this is benign and should be ignored.
	AlidnsErrCodeDomainRecordDuplicate = "DomainRecordDuplicate"
)

const (
	HeaderAction  = "X-Acs-Action"
	HeaderVersion = "X-Acs-Version"

	HeaderContentSha256  = "X-Acs-Content-Sha256"
	HeaderDate           = "X-Acs-Date"
	HeaderSignatureNonce = "X-Acs-Signature-Nonce"
	HeaderAuthorization  = "Authorization"
	HeaderSecurityToken  = "X-Acs-Security-Token"
)

const (
	SigAlgoACS3HMACSHA256 = "ACS3-HMAC-SHA256"
)

const (
	ContentTypeJSON         = "application/json"
	ContentTypeBinaryStream = "application/octet-stream"
	ContentTypeForm         = "application/x-www-form-urlencoded"
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
		do: &AliyunSignClient{
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
	return doAction[DescribeDomainsResponse](ctx, c, AlidnsActionDescribeDomains, req)
}

// DescribeDomainRecords — https://help.aliyun.com/document_detail/29774.html
func (c *Client) DescribeDomainRecords(ctx context.Context,
	req DescribeDomainRecordsRequest,
) (resp DescribeDomainRecordsResponse, err error) {
	defer c.metricsRouter.RecordAPI(opListRecords)(&err)
	return doAction[DescribeDomainRecordsResponse](ctx, c, AlidnsActionDescribeDomainRecords, req)
}

// AddDomainRecord — https://help.aliyun.com/document_detail/29772.html
func (c *Client) AddDomainRecord(ctx context.Context,
	req AddDomainRecordRequest,
) (resp AddDomainRecordResponse, err error) {
	defer c.metricsRouter.RecordAPI(opCreateRecord)(&err)
	return doAction[AddDomainRecordResponse](ctx, c, AlidnsActionAddDomainRecord, req)
}

// UpdateDomainRecord — https://help.aliyun.com/document_detail/29773.html
func (c *Client) UpdateDomainRecord(ctx context.Context,
	req UpdateDomainRecordRequest,
) (resp UpdateDomainRecordResponse, err error) {
	defer c.metricsRouter.RecordAPI(opUpdateRecord)(&err)
	return doAction[UpdateDomainRecordResponse](ctx, c, AlidnsActionUpdateDomainRecord, req)
}

// DeleteDomainRecord — https://help.aliyun.com/document_detail/29771.html
func (c *Client) DeleteDomainRecord(ctx context.Context,
	req DeleteDomainRecordRequest,
) (resp DeleteDomainRecordResponse, err error) {
	defer c.metricsRouter.RecordAPI(opDeleteRecord)(&err)
	return doAction[DeleteDomainRecordResponse](ctx, c, AlidnsActionDeleteDomainRecord, req)
}

// SearchDomain implements [ddnsx.DomainSearcher] by paging DescribeDomains.
// Each underlying HTTP call is recorded against the DescribeDomains metric.
//
// To minimise API calls the search walks suffixes of `search` (longest-first)
// and short-circuits as soon as one keyword turns up a domain that's a parent
// of `search`. Returns nil on transport / API failure so the cache treats it
// as "no result". Alidns keys zones by DomainName (no opaque id is needed by
// the record APIs), so the map's value equals its key.
func (c *Client) SearchDomain(ctx context.Context, search string) map[string]string {
	if mt.Done(ctx) || len(search) == 0 {
		return nil
	}

	const pageSize = 100

	logger := c.actionLogger(AlidnsActionDescribeDomains).
		With(zap.String("search", search))
	logger.Info("search domain id from upstream")

	result := make(map[string]string)
	for _, keyword := range append(domains.CutFromHead(search), "") {
		matched := false
		for pageNumber := 1; ; pageNumber++ {
			page, err := c.DescribeDomains(ctx, DescribeDomainsRequest{
				KeyWord:    keyword,
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})
			if err != nil {
				logger.Warn("list domains failed",
					zap.String("keyword", keyword), zap.Error(err))
				return nil
			}
			for _, di := range page.Domains.Domain {
				if !domains.IsDomainName(di.DomainName) {
					continue
				}
				result[di.DomainName] = di.DomainName
				if domains.IsSubDomain(search, di.DomainName) {
					matched = true
				}
			}
			if len(page.Domains.Domain) < pageSize ||
				int64(pageNumber)*pageSize >= page.TotalCount {
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

// doAction issues one POST against the Aliyun Alidns endpoint with action
// parameters carried as a JSON body. Alidns advertises support for JSON
// payloads via Content-Type: application/json under the V3 protocol — this
// keeps a single code path for both read and write actions and lets the
// httpx framework marshal the typed request struct. Metric recording is the
// caller's responsibility — each public API method defers
// metricsRouter.RecordAPI before invoking doAction so failure attribution
// stays at the action level.
func doAction[Resp any](ctx context.Context, c *Client, action string, body any) (Resp, error) {
	var zero Resp

	logger := c.actionLogger(action)

	req := newRequest(action)
	req.Body = body

	httpReq, err := req.ToRequestContext(ctx)
	if err != nil {
		return zero, fmt.Errorf("build request %s: %w", action, err)
	}

	// HTTP-level logging (transport error, status, duration) is recorded by
	// AliyunSignClient via httpx.HTTPRequestRecorder. We only log here for
	// concerns that the recorder can't see: body-decode failures and parsed
	// API errors.
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

	rawBody, err := freebuf.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("read body %s: %w", action, err)
	}
	defer rawBody.FreeMe()

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if jerr := json.NewDecoder(rawBody).Decode(apiErr); jerr != nil {
			logger.Warn("decode error response failed",
				zap.Int("status", resp.StatusCode), zap.Error(jerr))
			return zero, &httpx.BadStatusCodeError{Got: resp.StatusCode}
		}
		logger.Warn("api returned error",
			zap.String("error_code", apiErr.Code),
			zap.String("error_message", apiErr.Message),
			zap.String("request_id", apiErr.RequestID))
		return zero, apiErr
	}

	var out Resp
	if err := json.NewDecoder(rawBody).Decode(&out); err != nil {
		logger.Warn("decode response failed", zap.Error(err))
		return zero, fmt.Errorf("decode %s: %w", action, err)
	}
	return out, nil
}

func newRequest(action string) httpx.ReqConfig {
	r := httpx.NewReqConfig(http.MethodPost, AliyunDNSEndpoint)
	r.ExtendHeader.Set(HeaderAction, action)
	r.ExtendHeader.Set(HeaderVersion, AlidnsDefaultVersion)
	r.ExtendHeader.Set("Content-Type", ContentTypeJSON)
	return r
}
