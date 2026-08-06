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

const (
	APIEndpoint = "https://api.cloudflare.com/client/v4/zones"
)

var cloudflareAPIEndpoint = mt.Must(urlpkg.Parse(APIEndpoint))

type APIClient interface {
	ListZones(context.Context, ListZonesRequest) (ListZonesResponse, error)
	ListDNSRecords(context.Context, ListDNSRecordsRequest) (ListDNSRecordsResponse, error)
	CreateDNSRecord(context.Context, CreateDNSRecordRequest) (CreateDNSRecordResponse, error)
	UpdateDNSRecord(context.Context, UpdateDNSRecordRequest) (UpdateDNSRecordResponse, error)
	DeleteDNSRecord(context.Context, DeleteDNSRecordRequest) (DeleteDNSRecordResponse, error)
}

type defaultAPIClient struct {
	requester httpx.HTTPRequester
}

func NewAPIClient(logger *zap.Logger, requester httpx.HTTPRequester, token string) APIClient {
	logger = zaplog.DoNotPanic(logger).Named("api")
	return &defaultAPIClient{
		requester: &apiRequester{
			HTTPRequester: requester,
			Logger:        logger,
			Token:         token,
		},
	}
}

func (c *defaultAPIClient) ListZones(ctx context.Context, request ListZonesRequest) (ListZonesResponse, error) {
	req := c.newRequest(http.MethodGet)
	setQuery(req.Query, "status", string(request.Status))

	if len(request.Name) > 0 && len(strings.TrimSuffix(request.Name, ".")) > 0 {
		name := dns.Fqdn(request.Name)
		setQuery(req.Query, "name", "contains:"+name)
	}

	setPageQuery(req.Query, request.Page, request.PerPage)
	return doAPIRequest[ListZonesResponse](ctx, c, req)
}

func (c *defaultAPIClient) ListDNSRecords(
	ctx context.Context, request ListDNSRecordsRequest,
) (ListDNSRecordsResponse, error) {
	if request.ZoneID == "" {
		return mt.Zero[ListDNSRecordsResponse](), fmt.Errorf("ListDNSRecords: empty zone id")
	}

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
	if request.ZoneID == "" {
		return mt.Zero[CreateDNSRecordResponse](), fmt.Errorf("CreateDNSRecord: empty zone id")
	}

	req := c.newRequest(http.MethodPost)
	req.ExtendPath = append(req.ExtendPath, request.ZoneID, "dns_records")
	req.Body = request.Body
	return doAPIRequest[CreateDNSRecordResponse](ctx, c, req)
}

func (c *defaultAPIClient) UpdateDNSRecord(
	ctx context.Context, request UpdateDNSRecordRequest,
) (UpdateDNSRecordResponse, error) {
	if request.ZoneID == "" {
		return mt.Zero[UpdateDNSRecordResponse](), fmt.Errorf("UpdateDNSRecord: empty zone id")
	}

	if request.RecordID == "" {
		return mt.Zero[UpdateDNSRecordResponse](), fmt.Errorf("UpdateDNSRecord: empty record id")
	}

	req := c.newRequest(http.MethodPatch)
	req.ExtendPath = append(req.ExtendPath, request.ZoneID, "dns_records", request.RecordID)
	req.Body = request.Body
	return doAPIRequest[UpdateDNSRecordResponse](ctx, c, req)
}

func (c *defaultAPIClient) DeleteDNSRecord(
	ctx context.Context, request DeleteDNSRecordRequest,
) (DeleteDNSRecordResponse, error) {
	if request.ZoneID == "" {
		return mt.Zero[DeleteDNSRecordResponse](), fmt.Errorf("DeleteDNSRecord: empty zone id")
	}

	if request.RecordID == "" {
		return mt.Zero[DeleteDNSRecordResponse](), fmt.Errorf("DeleteDNSRecord: empty record id")
	}

	req := c.newRequest(http.MethodDelete)
	req.ExtendPath = append(req.ExtendPath, request.ZoneID, "dns_records", request.RecordID)
	return doAPIRequest[DeleteDNSRecordResponse](ctx, c, req)
}

func (c *defaultAPIClient) newRequest(method string) httpx.ReqConfig {
	return httpx.NewReqConfig(method, cloudflareAPIEndpoint)
}

func (c *defaultAPIClient) Do(request *http.Request) (*http.Response, error) {
	return c.requester.Do(request)
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
