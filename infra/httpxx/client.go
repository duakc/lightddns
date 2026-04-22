package httpxx

import (
	"net"
	"net/http"
	urlpkg "net/url"
	"time"

	"github.com/duakc/lightddns/infra/netool"

	"golang.org/x/net/http/httpproxy"
)

type Client struct {
	*http.Client

	HTTPRequester
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.HTTPRequester == nil {
		return c.Client.Do(req)
	}
	return c.HTTPRequester.Do(req)
}

func NewClient(opt ...HTTPClientOption) *Client {
	// modified from http.DefaultTransport
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	c := &Client{
		Client:        httpClient,
		HTTPRequester: httpClient,
	}
	for _, o := range opt {
		o.Apply(c)
	}

	return c
}

type HTTPClientOption interface {
	Apply(c *Client)
}

type FuncHTTPClientOption func(c *Client)

func (fn FuncHTTPClientOption) Apply(c *Client) {
	fn(c)
}

func ClientOptionWithToken(token string) HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		c.HTTPRequester = &TokenClient{
			HTTPRequester: c.HTTPRequester,
			Token:         token,
		}
	})
}

func ClientOptionWithRoundTripper(t http.RoundTripper) HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		c.Client.Transport = t
	})
}

func ClientOptionWithProxy(httpProxyURL, httpsProxyURL, noProxyURL string) HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		cfg := httpproxy.Config{
			HTTPProxy:  httpProxyURL,
			HTTPSProxy: httpsProxyURL,
			NoProxy:    noProxyURL,
		}
		c.Client.Transport.(*http.Transport).Proxy = func(req *http.Request) (*urlpkg.URL, error) {
			return cfg.ProxyFunc()(req.URL)
		}
	})
}

func ClientOptionEnableProxy() HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		c.Client.Transport.(*http.Transport).Proxy = http.ProxyFromEnvironment
	})
}

func ClientOptionWithDialer(d netool.Dialer) HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		c.Client.Transport.(*http.Transport).DialContext = d.DialContext
	})
}
