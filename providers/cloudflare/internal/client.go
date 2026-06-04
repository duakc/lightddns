package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"time"

	"github.com/duakc/lightddns/adapter/ddnsmetric"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/metrics"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

// Operation labels recorded against the provider API metrics vec. One label
// per HTTP request — pager Next() calls record one each.
const (
	opListZones      = "list_zones"
	opListDNSRecords = "list_dns_records"
	opCreateDNS      = "create_dns_record"
	opUpdateDNS      = "update_dns_record"
	opDeleteDNS      = "delete_dns_record"
)

var _ ddnsx.DomainIdFetcher = (*Client)(nil)

var cloudflareApiEndpoint = mt.Must(urlpkg.Parse("https://api.cloudflare.com/client/v4/zones"))

func NewClient(logger *zap.Logger, providerName string,
	do httpx.HTTPRequester, token string,
) *Client {
	return &Client{
		logger:       logger,
		providerName: providerName,
		do: &httpx.TokenClient{
			HTTPRequester: do,
			Token:         token,
		},
	}
}

type Client struct {
	logger       *zap.Logger
	providerName string
	do           httpx.HTTPRequester

	requestTotal    metrics.CounterVec
	requestFailures metrics.CounterVec
	requestDuration metrics.HistogramVec
}

// RegisterMetrics wires the client into the provider metric registry. Must be
// called once during the owning provider's PreStart, before any API method
// fires — otherwise recordAPICall hits nil vecs.
func (c *Client) RegisterMetrics(factory ddnsmetric.Factory) {
	labels := []string{constpkg.MetricLabelName, constpkg.MetricLabelOperation}
	c.requestTotal = factory.CounterVec(constpkg.MetricProviderRequestTotal,
		"Total provider API requests.", labels)
	c.requestFailures = factory.CounterVec(constpkg.MetricProviderRequestFailureTotal,
		"Failed provider API requests.", labels)
	c.requestDuration = factory.HistogramVec(constpkg.MetricProviderRequestDurationSeconds,
		"Provider API request duration.", labels, nil)
}

func (c *Client) recordAPICall(op string, start time.Time, err error) {
	c.requestTotal.With(c.providerName, op).Inc()
	c.requestDuration.With(c.providerName, op).Observe(time.Since(start).Seconds())
	if err != nil {
		c.requestFailures.With(c.providerName, op).Inc()
	}
}

// FetchDomainId implements [ddnsx.DomainIdFetcher]. It pages through all zones
// owned by the account and returns the full {zoneName -> zoneID} mapping.
// [ddnsx.DomainIdCache] picks the longest suffix match for the queried FQDN
// and remembers the rest for future lookups. Each underlying page request is
// itself recorded against the API metric — this method does not record a
// separate top-level metric. Returns nil on transport / API failure so the
// cache treats it as "no result".
func (c *Client) FetchDomainId(ctx context.Context, domain string) map[string]string {
	if mt.Done(ctx) || len(domain) == 0 {
		return nil
	}
	logger := c.actionLogger(opListZones).With(zap.String("domain", domain))
	logger.Info("search zone id from upstream")

	pager := c.ListZones()
	result := make(map[string]string)
	for page, perr := pager.Next(ctx); perr != io.EOF; page, perr = pager.Next(ctx) {
		if perr != nil {
			logger.Warn("cloudflare: list zones failed", zap.Error(perr))
			return nil
		}
		for i := 0; i < len(page); i++ {
			zone := page[i]
			if !domains.IsDomainName(zone.Name) {
				logger.Warn("cloudflare: upstream returned a bad zone name",
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
	return NewPaging[Zone](c, opListZones, r)
}

func (c *Client) ListZoneName(name string) *PageConfig[Zone] {
	r := c.NewRequestConfig(http.MethodGet)
	r.Query.Set("status", "active")
	r.Query.Set("name", name)
	return NewPaging[Zone](c, opListZones, r)
}

func (c *Client) ListDNSRecords(name string, zoneID string, qtype string) (*PageConfig[DNSRecord], error) {
	r := c.NewRequestConfig(http.MethodGet)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records")
	r.Query.Set("name", name)
	r.Query.Set("type", qtype)
	return NewPaging[DNSRecord](c, opListDNSRecords, r), nil
}

func (c *Client) CreateDNSRecords(ctx context.Context, zoneID string,
	content UpdateDNSRecordRequest,
) (err error) {
	start := time.Now()
	defer func() { c.recordAPICall(opCreateDNS, start, err) }()

	logger := c.actionLogger(opCreateDNS).With(zap.String("zone_id", zoneID))
	logger.Debug("cloudflare: api call start")
	r := c.NewRequestConfig(http.MethodPost)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records")
	r.Body = content
	createResult, response, perr := httpx.JSONRequest[Response](ctx, c.do, r, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	if joined := createResult.JoinError(perr); joined != nil {
		logger.Warn("cloudflare: create dns record failed", zap.Error(joined))
		err = joined
		return err
	}
	return nil
}

func (c *Client) UpdateDNSRecords(ctx context.Context, zoneID string, recordID string,
	content UpdateDNSRecordRequest,
) (err error) {
	start := time.Now()
	defer func() { c.recordAPICall(opUpdateDNS, start, err) }()

	logger := c.actionLogger(opUpdateDNS).With(
		zap.String("zone_id", zoneID),
		zap.String("record_id", recordID),
	)
	logger.Debug("cloudflare: api call start")
	r := c.NewRequestConfig(http.MethodPatch)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records", recordID)
	r.Body = content
	updateResult, response, perr := httpx.JSONRequest[Response](ctx, c.do, r, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	if joined := updateResult.JoinError(perr); joined != nil {
		logger.Warn("cloudflare: update dns record failed", zap.Error(joined))
		err = joined
		return err
	}
	return nil
}

func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID string, dnsRecordID string) (err error) {
	start := time.Now()
	defer func() { c.recordAPICall(opDeleteDNS, start, err) }()

	logger := c.actionLogger(opDeleteDNS).With(
		zap.String("zone_id", zoneID),
		zap.String("record_id", dnsRecordID),
	)
	logger.Debug("cloudflare: api call start")
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
		logger.Warn("cloudflare: delete request failed", zap.Error(perr))
		err = &httpx.BaseResponseError{Err: perr, Method: r.Method}
		return err
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		logger.Warn("cloudflare: delete bad status", zap.Int("status", resp.StatusCode))
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
