package internal

import (
	"context"
	"fmt"
	"io"
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

const (
	opDescribeDomains = ddnsprovider.OpDescribeDomains
	opListRecords     = ddnsprovider.OpListRecords
	opCreateRecord    = ddnsprovider.OpCreateRecord
	opUpdateRecord    = ddnsprovider.OpUpdateRecord
	opDeleteRecord    = ddnsprovider.OpDeleteRecord
)

var _ ddnsx.DomainSearcher = (*Client)(nil)

var cloudflareApiEndpoint = mt.Must(urlpkg.Parse("https://api.cloudflare.com/client/v4/zones"))

func NewClient(ctx context.Context, logger *zap.Logger, providerName string,
	do httpx.HTTPRequester, token string,
) *Client {
	router := ddnsprovider.NewMetricsRouter(
		services.Lookup[ddnsmetric.ProviderFactory](ctx), providerName)
	router.RegisterDefault()

	return &Client{
		logger:        logger,
		metricsRouter: router,
		do: &CloudflareHTTPRequester{
			HTTPRequester: do,
			Logger:        logger,
			Token:         token,
		},
	}
}

type Client struct {
	logger *zap.Logger
	do     httpx.HTTPRequester

	metricsRouter *ddnsprovider.ApiMetricsRouter
}

func (c *Client) SearchDomain(ctx context.Context, search string) map[string]string {
	logger := c.actionLogger(opDescribeDomains)
	logger.Info("search zone id from upstream")

	pager := c.ListZoneName(search)
	result := make(map[string]string)

	for page, perr := pager.Next(ctx); perr != io.EOF; page, perr = pager.Next(ctx) {
		if perr != nil {
			logger.Warn("list zones failed", zap.Error(perr))
			return nil
		}
		for i := 0; i < len(page); i++ {
			zone := page[i]
			if !domains.IsDomainName(zone.Name) {
				logger.Warn("upstream returned a bad zone name",
					zap.String("zone_name", zone.Name))
				continue
			}
			result[zone.Name] = zone.ID
		}
	}
	return result
}

func (c *Client) NewRequestConfig(method string) httpx.ReqConfig {
	return httpx.NewReqConfig(method, cloudflareApiEndpoint)
}

func (c *Client) ListZones() *PageConfig[Zone] {
	r := c.NewRequestConfig(http.MethodGet)
	r.Query.Set("status", "active")
	return NewPaging[Zone](c, opDescribeDomains, r)
}

func (c *Client) ListZoneName(name string) *PageConfig[Zone] {
	r := c.NewRequestConfig(http.MethodGet)
	r.Query.Set("status", "active")
	r.Query.Set("name", "contains:"+name)
	return NewPaging[Zone](c, opDescribeDomains, r)
}

func (c *Client) ListDNSRecords(name string, zoneID string, qtype string) (*PageConfig[DNSRecord], error) {
	r := c.NewRequestConfig(http.MethodGet)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records")
	r.Query.Set("name", name)
	r.Query.Set("type", qtype)
	return NewPaging[DNSRecord](c, opListRecords, r), nil
}

func (c *Client) CreateDNSRecords(ctx context.Context, zoneID string,
	content UpdateDNSRecordRequest,
) (err error) {
	defer c.metricsRouter.RecordAPI(opCreateRecord)(&err)

	logger := c.actionLogger(opCreateRecord).With(zap.String("zone_id", zoneID))
	logger.Debug("api call start")
	r := c.NewRequestConfig(http.MethodPost)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records")
	r.Body = content
	createResult, response, perr := httpx.JSONRequest[Response](ctx, c.do, r, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	if joined := createResult.JoinError(perr); joined != nil {
		logger.Warn("create dns record failed", zap.Error(joined))
		err = joined
		return err
	}

	return nil
}

func (c *Client) UpdateDNSRecords(ctx context.Context, zoneID string, recordID string,
	content UpdateDNSRecordRequest,
) (err error) {
	defer c.metricsRouter.RecordAPI(opUpdateRecord)(&err)

	logger := c.actionLogger(opUpdateRecord).With(
		zap.String("zone_id", zoneID),
		zap.String("record_id", recordID),
	)

	r := c.NewRequestConfig(http.MethodPatch)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records", recordID)
	r.Body = content
	updateResult, response, perr := httpx.JSONRequest[Response](ctx, c.do, r, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	if joined := updateResult.JoinError(perr); joined != nil {
		logger.Warn("update dns record failed", zap.Error(joined))
		err = joined
		return err
	}
	return nil
}

func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID string, dnsRecordID string) (err error) {
	defer c.metricsRouter.RecordAPI(opDeleteRecord)(&err)

	logger := c.actionLogger(opDeleteRecord).With(
		zap.String("zone_id", zoneID),
		zap.String("record_id", dnsRecordID),
	)

	r := c.NewRequestConfig(http.MethodDelete)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records", dnsRecordID)
	httpReq, perr := r.ToRequestContext(ctx)
	if perr != nil {
		err = fmt.Errorf("build request: %w", perr)
		return err
	}

	resp, perr := c.do.Do(httpReq)
	if resp != nil {
		defer resp.Body.Close()
	}
	if perr != nil {
		logger.Warn("delete request failed", zap.Error(perr))
		err = &httpx.BaseResponseError{Err: perr, Method: r.Method}
		return err
	}

	if resp != nil && resp.StatusCode != http.StatusOK {
		logger.Warn("delete bad status", zap.Int("status", resp.StatusCode))
		err = &httpx.BadStatusCodeError{
			Got: resp.StatusCode,
		}
		return err
	}
	return nil
}

func (c *Client) actionLogger(action string) *zap.Logger {
	return c.logger.With(zap.String("action", action))
}
