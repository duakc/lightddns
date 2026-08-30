package tencentcloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/infra/netx/httpx"
	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/mt"

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
	// domain exists but has no records matching the filter.
	DNSPodErrCodeNoDataOfRecord = "ResourceNotFound.NoDataOfRecord"

	// DNSPodErrCodeMustAddDefaultLineFirst is returned when a non-default line
	// is requested before the default line has been created.
	DNSPodErrCodeMustAddDefaultLineFirst = "FailedOperation.MustAddDefaultLineFirst"

	DNSPodErrorCodeURL = "https://cloud.tencent.com/document/api/1427/56188"
)

const (
	HeaderAction    = "X-TC-Action"
	HeaderVersion   = "X-TC-Version"
	HeaderTimestamp = "X-TC-Timestamp"
	HeaderToken     = "X-TC-Token"
	HeaderLanguage  = "X-TC-Language"
	HeaderRegion    = "X-TC-Region"
)

const (
	ContentTypeJSON = "application/json; charset=utf-8"
)

const (
	DNSPodEndpoint = "https://dnspod.tencentcloudapi.com"
)

var dnsPodEndpointUrl = mt.Must(urlpkg.Parse(DNSPodEndpoint))

type APIClient interface {
	DescribeDomainFilterList(context.Context, DescribeDomainFilterListRequest) (DescribeDomainFilterListResponse, error)
	DescribeRecordList(context.Context, DescribeRecordListRequest) (DescribeRecordListResponse, error)
	CreateRecord(context.Context, CreateRecordRequest) (CreateRecordResponse, error)
	ModifyRecord(context.Context, ModifyRecordRequest) (ModifyRecordResponse, error)
	DeleteRecord(context.Context, DeleteRecordRequest) (DeleteRecordResponse, error)
}

// defaultAPIClient implements the DNSPod RPC API over HTTP with TC3 signing.
type defaultAPIClient struct {
	logger    *zap.Logger
	requester httpx.HTTPRequester
}

func NewAPIClient(logger *zap.Logger, requester httpx.HTTPRequester, secretId, secretKey string) APIClient {
	logger = zaplog.DoNotPanic(logger).Named("api")
	return &defaultAPIClient{
		logger: logger,
		requester: &TencentSignHTTPRequester{
			HTTPRequester: requester,
			Logger:        logger,
			SecretId:      secretId,
			SecretKey:     secretKey,
			Service:       DNSPodServiceName,
		},
	}
}

// DescribeDomainFilterList — https://cloud.tencent.com/document/api/1427/56173
func (c *defaultAPIClient) DescribeDomainFilterList(ctx context.Context,
	req DescribeDomainFilterListRequest,
) (resp DescribeDomainFilterListResponse, err error) {
	return doAction[DescribeDomainFilterListResponse](ctx, c, DNSPodActionDescribeDomainFilterList, req)
}

// DescribeRecordList — https://cloud.tencent.com/document/api/1427/56166
func (c *defaultAPIClient) DescribeRecordList(ctx context.Context,
	req DescribeRecordListRequest,
) (resp DescribeRecordListResponse, err error) {
	return doAction[DescribeRecordListResponse](ctx, c, DNSPodActionDescribeRecordList, req)
}

// CreateRecord — https://cloud.tencent.com/document/api/1427/56180
func (c *defaultAPIClient) CreateRecord(ctx context.Context,
	req CreateRecordRequest,
) (resp CreateRecordResponse, err error) {
	resp, err = doAction[CreateRecordResponse](ctx, c, DNSPodActionCreateRecord, req)
	if apiErr, ok := errors.AsType[*APIError](err); ok && apiErr.Code == DNSPodErrCodeMustAddDefaultLineFirst {
		err = fmt.Errorf("%w: Tencent Cloud rejected line %q (%s: %s). This can happen when the %q line record does not exist. Add %q to the provider's \"lines\" configuration and create/update that default record first; otherwise the default record will not be updated on later startups: %w",
			ErrDefaultRecordLineRequired, req.RecordLine, apiErr.Code, apiErr.Message,
			DefaultRecordLine, DefaultRecordLine, err)
	}
	return resp, err
}

// ModifyRecord — https://cloud.tencent.com/document/api/1427/56157
func (c *defaultAPIClient) ModifyRecord(ctx context.Context,
	req ModifyRecordRequest,
) (resp ModifyRecordResponse, err error) {
	resp, err = doAction[ModifyRecordResponse](ctx, c, DNSPodActionModifyRecord, req)
	if apiErr, ok := errors.AsType[*APIError](err); ok && apiErr.Code == DNSPodErrCodeMustAddDefaultLineFirst {
		err = fmt.Errorf("%w: Tencent Cloud rejected line %q (%s: %s). This can happen when the %q line record does not exist. Add %q to the provider's \"lines\" configuration and create/update that default record first; otherwise the default record will not be updated on later startups: %w",
			ErrDefaultRecordLineRequired, req.RecordLine, apiErr.Code, apiErr.Message,
			DefaultRecordLine, DefaultRecordLine, err)
	}
	return resp, err
}

// DeleteRecord — https://cloud.tencent.com/document/api/1427/56176
func (c *defaultAPIClient) DeleteRecord(ctx context.Context,
	req DeleteRecordRequest,
) (resp DeleteRecordResponse, err error) {
	return doAction[DeleteRecordResponse](ctx, c, DNSPodActionDeleteRecord, req)
}

func doAction[Resp any](ctx context.Context, c *defaultAPIClient, action string, body any) (Resp, error) {
	var zero Resp

	logger := c.logger.With(zap.String("action", action))

	req := newRequest(http.MethodPost, action)
	req.ExtendHeader.Set(httpx.HeaderHost, dnsPodEndpointUrl.Host)
	req.ExtendHeader.Set(httpx.HeaderContentType, ContentTypeJSON)

	req.Body = body

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
		var out Response[Resp]
		if decodeErr := json.NewDecoder(resp.Body).Decode(&out); decodeErr == nil && out.Error != nil {
			out.Error.StatusCode = resp.StatusCode
			logger.Warn("api returned error",
				zap.String("error_code", out.Error.Code),
				zap.String("error_message", out.Error.Message),
				zap.String("request_id", out.RequestID))
			return zero, out.Error
		}
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
	req := httpx.NewReqConfig(method, dnsPodEndpointUrl)
	req.ExtendHeader.Set(HeaderAction, action)
	req.ExtendHeader.Set(HeaderVersion, DNSPodDefaultVersion)
	return req
}
