package aliyun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/infra/netx/httpx"
	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/mt"

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

	// AlidnsErrCodeInvalidParameterLine is returned when the requested line
	// cannot be used for the record (for example, before the default line has
	// been created).
	AlidnsErrCodeInvalidParameterLine = "InvalidParameter.Line"

	AlidnsErrorCodeURL = "https://api.aliyun.com/document/Alidns/2015-01-09/errorCode"
)

const (
	DNSAPIEndpoint = "https://alidns.aliyuncs.com"
)

var aliyunDNSEndpoint = mt.Must(urlpkg.Parse(DNSAPIEndpoint))

type APIClient interface {
	DescribeDomains(context.Context, DescribeDomainsRequest) (DescribeDomainsResponse, error)
	DescribeDomainRecords(context.Context, DescribeDomainRecordsRequest) (DescribeDomainRecordsResponse, error)
	AddDomainRecord(context.Context, AddDomainRecordRequest) (AddDomainRecordResponse, error)
	UpdateDomainRecord(context.Context, UpdateDomainRecordRequest) (UpdateDomainRecordResponse, error)
	DeleteDomainRecord(context.Context, DeleteDomainRecordRequest) (DeleteDomainRecordResponse, error)
}

// defaultAPIClient performs one signed Aliyun RPC request per method call.
type defaultAPIClient struct {
	logger    *zap.Logger
	requester httpx.HTTPRequester
}

func NewAPIClient(logger *zap.Logger, requester httpx.HTTPRequester,
	secretAccessKeyId, secretAccessKeySecret, secretSecurityToken string,
) APIClient {
	logger = zaplog.DoNotPanic(logger).Named("api")
	return &defaultAPIClient{
		logger: logger,
		requester: &RpcSignRequester{
			HTTPRequester:         requester,
			Logger:                logger,
			SecretSecurityToken:   secretSecurityToken,
			SecretAccessKeyId:     secretAccessKeyId,
			SecretAccessKeySecret: secretAccessKeySecret,
		},
	}
}

// DescribeDomains — https://help.aliyun.com/document_detail/29751.html
func (c *defaultAPIClient) DescribeDomains(ctx context.Context,
	req DescribeDomainsRequest,
) (resp DescribeDomainsResponse, err error) {
	return doAction[DescribeDomainsResponse](ctx, c, AlidnsActionDescribeDomains, req.Query())
}

// DescribeDomainRecords — https://help.aliyun.com/document_detail/29774.html
func (c *defaultAPIClient) DescribeDomainRecords(ctx context.Context,
	req DescribeDomainRecordsRequest,
) (resp DescribeDomainRecordsResponse, err error) {
	return doAction[DescribeDomainRecordsResponse](ctx, c, AlidnsActionDescribeDomainRecords, req.Query())
}

// AddDomainRecord — https://help.aliyun.com/document_detail/29772.html
func (c *defaultAPIClient) AddDomainRecord(ctx context.Context,
	req AddDomainRecordRequest,
) (resp AddDomainRecordResponse, err error) {
	resp, err = doAction[AddDomainRecordResponse](ctx, c, AlidnsActionAddDomainRecord, req.Query())
	if apiErr, ok := errors.AsType[*APIError](err); ok && apiErr.Code == AlidnsErrCodeInvalidParameterLine {
		err = fmt.Errorf("%w: Aliyun rejected line %q (%s: %s). This can happen when the %q line record does not exist. Add %q to the provider's \"lines\" configuration and create/update that default record first; otherwise the default record will not be updated on later startups: %w",
			ErrDefaultRecordLineRequired, req.Line, apiErr.Code, apiErr.Message,
			DefaultRecordLine, DefaultRecordLine, err)
	}
	return resp, err
}

// UpdateDomainRecord — https://help.aliyun.com/document_detail/29773.html
func (c *defaultAPIClient) UpdateDomainRecord(ctx context.Context,
	req UpdateDomainRecordRequest,
) (resp UpdateDomainRecordResponse, err error) {
	resp, err = doAction[UpdateDomainRecordResponse](ctx, c, AlidnsActionUpdateDomainRecord, req.Query())
	if apiErr, ok := errors.AsType[*APIError](err); ok && apiErr.Code == AlidnsErrCodeInvalidParameterLine {
		err = fmt.Errorf("%w: Aliyun rejected line %q (%s: %s). This can happen when the %q line record does not exist. Add %q to the provider's \"lines\" configuration and create/update that default record first; otherwise the default record will not be updated on later startups: %w",
			ErrDefaultRecordLineRequired, req.Line, apiErr.Code, apiErr.Message,
			DefaultRecordLine, DefaultRecordLine, err)
	}
	return resp, err
}

// DeleteDomainRecord — https://help.aliyun.com/document_detail/29771.html
func (c *defaultAPIClient) DeleteDomainRecord(ctx context.Context,
	req DeleteDomainRecordRequest,
) (resp DeleteDomainRecordResponse, err error) {
	return doAction[DeleteDomainRecordResponse](ctx, c, AlidnsActionDeleteDomainRecord, req.Query())
}

func doAction[Resp any](ctx context.Context, c *defaultAPIClient, action string, query urlpkg.Values) (Resp, error) {
	var zero Resp

	logger := c.logger.WithLazy(
		zap.String("action", action),
	)

	req := newRequest(action)
	maps.Copy(req.Query, query)

	httpReq, err := req.ToRequestContext(ctx)
	if err != nil {
		return zero, fmt.Errorf("build request %s: %w", action, err)
	}

	resp, err := c.requester.Do(httpReq)
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
		logger.Warn("decode response failed", zap.Error(err))
		return zero, fmt.Errorf("decode %s: %w", action, err)
	}
	return out, nil
}

func newRequest(action string) httpx.ReqConfig {
	r := httpx.NewReqConfig(http.MethodPost, aliyunDNSEndpoint)
	r.Query.Set(ParamAction, action)
	r.Query.Set(ParamVersion, AlidnsDefaultVersion)
	return r
}
