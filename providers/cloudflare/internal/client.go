package internal

import (
	"context"
	"fmt"
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/infra/netool/httpx"
	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

var cloudflareApiEndpoint = mt.Must(urlpkg.Parse("https://api.cloudflare.com/client/v4/zones"))

type Client struct {
	client httpx.HTTPRequester
}

func NewClient(client httpx.HTTPRequester, token string) *Client {
	return &Client{
		client: &httpx.TokenClient{
			HTTPRequester: client,
			Token:         token,
		},
	}
}

func actionLogger(ctx context.Context, action string) *zap.Logger {
	return zaplog.FromOrPackage(ctx, "cloudflare", "internal").
		With(zap.String("action", action))
}

func (c *Client) NewRequestConfig(method string) httpx.ReqConfig {
	return httpx.NewReqConfig(method, cloudflareApiEndpoint)
}

func (c *Client) ListZones() *PageConfig[Zone] {
	r := c.NewRequestConfig(http.MethodGet)
	r.Query.Set("status", "active")
	return NewPaging[Zone](r, c.client)
}

func (c *Client) ListZoneName(name string) *PageConfig[Zone] {
	r := c.NewRequestConfig(http.MethodGet)
	r.Query.Set("status", "active")
	r.Query.Set("name", name)
	return NewPaging[Zone](r, c.client)
}

func (c *Client) ListDNSRecords(name string, zoneID string, qtype string) (*PageConfig[DNSRecord], error) {
	r := c.NewRequestConfig(http.MethodGet)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records")
	r.Query.Set("name", name)
	r.Query.Set("type", qtype)
	return NewPaging[DNSRecord](r, c.client), nil
}

func (c *Client) CreateDNSRecords(ctx context.Context, zoneID string,
	content UpdateDNSRecordRequest,
) error {
	logger := actionLogger(ctx, "create_dns_record").With(zap.String("zone_id", zoneID))
	logger.Debug("cloudflare: api call start")
	r := c.NewRequestConfig(http.MethodPost)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records")
	r.Body = content
	createResult, response, err := httpx.JSONRequest[Response](ctx, c.client, r, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	if joined := createResult.JoinError(err); joined != nil {
		logger.Warn("cloudflare: create dns record failed", zap.Error(joined))
		return joined
	}
	return nil
}

func (c *Client) UpdateDNSRecords(ctx context.Context, zoneID string, recordID string,
	content UpdateDNSRecordRequest,
) error {
	logger := actionLogger(ctx, "update_dns_record").
		With(zap.String("zone_id", zoneID), zap.String("record_id", recordID))
	logger.Debug("cloudflare: api call start")
	r := c.NewRequestConfig(http.MethodPatch)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records", recordID)
	r.Body = content
	createResult, response, err := httpx.JSONRequest[Response](ctx, c.client, r, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	if joined := createResult.JoinError(err); joined != nil {
		logger.Warn("cloudflare: update dns record failed", zap.Error(joined))
		return joined
	}
	return nil
}

func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID string, dnsRecordID string) error {
	logger := actionLogger(ctx, "delete_dns_record").
		With(zap.String("zone_id", zoneID), zap.String("record_id", dnsRecordID))
	logger.Debug("cloudflare: api call start")
	r := c.NewRequestConfig(http.MethodDelete)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records", dnsRecordID)
	httpReq, err := r.ToRequestContext(ctx)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		logger.Warn("cloudflare: delete request failed", zap.Error(err))
		return &httpx.BaseResponseError{Err: err, Method: r.Method}
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		logger.Warn("cloudflare: delete bad status", zap.Int("status", resp.StatusCode))
		return &httpx.BadStatusCodeError{
			Got: resp.StatusCode,
		}
	}
	return nil
}
