package internal

import (
	"context"
	"fmt"
	"net/http"
	urlpkg "net/url"

	"github.com/duakc/lightddns/infra/netool/httpx"

	"github.com/duakc/mt"
)

var cloudflareApiEndpoint = mt.Must(urlpkg.Parse("https://api.cloudflare.com/client/v4/zones"))

type Client struct {
	client httpx.HTTPRequester
}

func NewClient(ctx context.Context, client httpx.HTTPRequester) *Client {
	return &Client{
		client: client,
	}
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
	r := c.NewRequestConfig(http.MethodPost)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records")
	r.Body = content
	createResult, response, err := httpx.JSONRequest[Response](ctx, c.client, r, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	return createResult.JoinError(err)
}

func (c *Client) UpdateDNSRecords(ctx context.Context, zoneID string, recordID string,
	content UpdateDNSRecordRequest,
) error {
	r := c.NewRequestConfig(http.MethodPatch)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records", recordID)
	r.Body = content
	createResult, response, err := httpx.JSONRequest[Response](ctx, c.client, r, httpx.RespPolicy{})
	if response != nil {
		defer response.Body.Close()
	}

	return createResult.JoinError(err)
}

func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID string, dnsRecordID string) error {
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
		return &httpx.BaseResponseError{Err: err, Method: r.Method}
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		return &httpx.BadStatusCodeError{
			Got: resp.StatusCode,
		}
	}
	return nil
}
