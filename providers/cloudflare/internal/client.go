package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/duakc/lightddns/infra/httpxx"
)

const cloudflareApiEndpoint = "https://api.cloudflare.com/client/v4/zones"

type Client struct {
	client httpxx.HTTPRequester
}

func NewClient(ctx context.Context, token string) *Client {
	return &Client{
		client: &httpxx.ValidClient{HTTPRequester: &httpxx.TokenClient{
			HTTPRequester: http.DefaultClient,
			Token:         token,
		}},
	}
}

func (c *Client) NewRequestConfig(method string) httpxx.ReqConfig {
	return httpxx.ReqConfig{
		Method:  method,
		BaseURL: cloudflareApiEndpoint,
	}
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
	content UpdateDNSRecordRequest) error {
	r := c.NewRequestConfig(http.MethodPost)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records")
	createResult, response, err := httpxx.JSONRequest[Response](ctx, c.client, r, content)
	if response != nil {
		defer response.Body.Close()
	}

	return createResult.JoinError(err)
}

func (c *Client) UpdateDNSRecords(ctx context.Context, zoneID string, recordID string,
	content UpdateDNSRecordRequest) error {
	r := c.NewRequestConfig(http.MethodPatch)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_records", recordID)
	createResult, response, err := httpxx.JSONRequest[Response](ctx, c.client, r, content)
	if response != nil {
		defer response.Body.Close()
	}

	return createResult.JoinError(err)
}

func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID string, dnsRecordID string) error {
	r := c.NewRequestConfig(http.MethodDelete)
	r.ExtendPath = append(r.ExtendPath, zoneID, "dns_record", dnsRecordID)
	httpReq, err := r.ToRequestContext(ctx)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return &httpxx.ResponseError{Err: err, URL: httpReq.URL.String(), Method: r.Method}
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		return &httpxx.BadStatusCodeError{
			Got: resp.StatusCode,
		}
	}
	return nil
}
