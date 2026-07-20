package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strconv"
	"strings"

	"github.com/duakc/lightddns/infra/netx/httpx"
	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/mt"

	"github.com/miekg/dns"
	"go.uber.org/zap"
)

var cloudflareAPIEndpoint = mt.Must(urlpkg.Parse("https://api.cloudflare.com/client/v4/zones"))

type APIClient interface {
	ListZones(context.Context, ListZonesRequest) (ListZonesResponse, error)
	ListDNSRecords(context.Context, ListDNSRecordsRequest) (ListDNSRecordsResponse, error)
	CreateDNSRecord(context.Context, CreateDNSRecordRequest) (CreateDNSRecordResponse, error)
	UpdateDNSRecord(context.Context, UpdateDNSRecordRequest) (UpdateDNSRecordResponse, error)
	DeleteDNSRecord(context.Context, DeleteDNSRecordRequest) (DeleteDNSRecordResponse, error)
}

type defaultAPIClient struct {
	do httpx.HTTPRequester
}

func NewAPIClient(logger *zap.Logger, do httpx.HTTPRequester, token string) APIClient {
	logger = zaplog.DoNotPanic(logger).Named("api")
	return &defaultAPIClient{do: &apiRequester{
		HTTPRequester: do,
		Logger:        logger,
		Token:         token,
	}}
}

func (c *defaultAPIClient) ListZones(ctx context.Context, request ListZonesRequest) (ListZonesResponse, error) {
	req := c.newRequest(http.MethodGet)
	setQuery(req.Query, "status", request.Status)
	if len(request.Name) > 0 && strings.TrimSuffix(request.Name, ".") != "" {
		name := dns.Fqdn(request.Name)
		setQuery(req.Query, "name", "contains:"+name)
	}

	setPageQuery(req.Query, request.Page, request.PerPage)
	return doAPIRequest[ListZonesResponse](ctx, c, req)
}

func (c *defaultAPIClient) ListDNSRecords(
	ctx context.Context, request ListDNSRecordsRequest,
) (ListDNSRecordsResponse, error) {
	req := c.newRequest(http.MethodGet)
	req.ExtendPath = append(req.ExtendPath, request.ZoneID, "dns_records")
	setQuery(req.Query, "name", request.Name)
	setQuery(req.Query, "type", request.Type)
	setPageQuery(req.Query, request.Page, request.PerPage)
	return doAPIRequest[ListDNSRecordsResponse](ctx, c, req)
}

func (c *defaultAPIClient) CreateDNSRecord(
	ctx context.Context, request CreateDNSRecordRequest,
) (CreateDNSRecordResponse, error) {
	req := c.newRequest(http.MethodPost)
	req.ExtendPath = append(req.ExtendPath, request.ZoneID, "dns_records")
	req.Body = request.Body
	return doAPIRequest[CreateDNSRecordResponse](ctx, c, req)
}

func (c *defaultAPIClient) UpdateDNSRecord(
	ctx context.Context, request UpdateDNSRecordRequest,
) (UpdateDNSRecordResponse, error) {
	req := c.newRequest(http.MethodPatch)
	req.ExtendPath = append(req.ExtendPath, request.ZoneID, "dns_records", request.RecordID)
	req.Body = request.Body
	return doAPIRequest[UpdateDNSRecordResponse](ctx, c, req)
}

func (c *defaultAPIClient) DeleteDNSRecord(
	ctx context.Context, request DeleteDNSRecordRequest,
) (DeleteDNSRecordResponse, error) {
	req := c.newRequest(http.MethodDelete)
	req.ExtendPath = append(req.ExtendPath, request.ZoneID, "dns_records", request.RecordID)
	return doAPIRequest[DeleteDNSRecordResponse](ctx, c, req)
}

func (c *defaultAPIClient) newRequest(method string) httpx.ReqConfig {
	return httpx.NewReqConfig(method, cloudflareAPIEndpoint)
}

func (c *defaultAPIClient) Do(request *http.Request) (*http.Response, error) {
	return c.do.Do(request)
}

func doAPIRequest[Resp interface{ JoinError(error) error }](
	ctx context.Context, client *defaultAPIClient, request httpx.ReqConfig,
) (Resp, error) {
	result, response, requestErr := httpx.JSONRequest[Resp](
		ctx, client, request, httpx.RespPolicy{},
	)
	if response != nil {
		defer response.Body.Close()
	}
	if err := result.JoinError(requestErr); err != nil {
		return result, fmt.Errorf("cloudflare API: %w", err)
	}
	return result, nil
}

func setQuery(query urlpkg.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}

func setPageQuery(query urlpkg.Values, page, perPage int) {
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		query.Set("per_page", strconv.Itoa(perPage))
	}
}
