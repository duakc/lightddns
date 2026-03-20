package internal

import (
	"context"
	"net/http"

	"github.com/duakc/lightddns/infra/httpxx"
)

const cloudflareApiEndpoint = "https://api.cloudflare.com/client/v4/zones"

type Client struct {
	client *httpxx.TokenClient
}

func NewClient(ctx context.Context, token string) *Client {
	return &Client{
		client: &httpxx.TokenClient{
			HTTPRequester: http.DefaultClient,
			Token:         token,
		},
	}
}

func (c *Client) ListZoneName(name string) *PageConfig[Zone] {
	r := c.NewRequestConfig(http.MethodGet)
	r.Query.Set("status", "active")
	r.Query.Set("name", name)
	return NewPaging[Zone](r, c.client)
}

func (c *Client) NewRequestConfig(method string) httpxx.ReqConfig {
	return httpxx.ReqConfig{
		Method:  method,
		BaseURL: cloudflareApiEndpoint,
	}
}
