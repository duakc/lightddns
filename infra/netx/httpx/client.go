package httpx

import (
	"net/http"
	urlpkg "net/url"
	"time"

	"github.com/duakc/lightddns/infra/netx/dialerx"

	"go.uber.org/zap"
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

func NewClient(dialer dialerx.Dialer, opt ...HTTPClientOption) *Client {
	// modified from http.DefaultTransport
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
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

// ClientOptionWithProxy configures the proxy on the underlying *http.Transport.
// It is a no-op if a custom RoundTripper has replaced the default transport
// (see ClientOptionWithRoundTripper).
func ClientOptionWithProxy(httpProxyURL, httpsProxyURL, noProxyURL string) HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		t := getClientTransport(c)
		cfg := httpproxy.Config{
			HTTPProxy:  httpProxyURL,
			HTTPSProxy: httpsProxyURL,
			NoProxy:    noProxyURL,
		}
		t.Proxy = func(req *http.Request) (*urlpkg.URL, error) {
			return cfg.ProxyFunc()(req.URL)
		}
	})
}

// ClientOptionEnableProxy enables proxy lookup from the environment on the
// underlying *http.Transport. It is a no-op if a custom RoundTripper has
// replaced the default transport (see ClientOptionWithRoundTripper).
func ClientOptionEnableProxy() HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		t := getClientTransport(c)
		t.Proxy = http.ProxyFromEnvironment
	})
}

func ClientOptionWithHeader(key, value string) HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		if len(key) == 0 { // empty value is allowed.
			return
		}

		if headerRequester, isHeaderRequester := c.HTTPRequester.(*HeaderClient); isHeaderRequester {
			headerRequester.Headers.Add(key, value)
			return
		}
		c.HTTPRequester = &HeaderClient{
			HTTPRequester: c.HTTPRequester,
			Headers: http.Header{
				key: []string{value},
			},
		}
	})
}

func ClientOptionWithHeaders(header http.Header) HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		if len(header) == 0 {
			return
		}

		if headerRequester, isHeaderRequester := c.HTTPRequester.(*HeaderClient); isHeaderRequester {
			ExtendHeaders(headerRequester.Headers, header)
			return
		}
		c.HTTPRequester = &HeaderClient{
			HTTPRequester: c.HTTPRequester,
			Headers:       header,
		}
	})
}

func ClientOptionWithDebug() HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		c.HTTPRequester = &DebugClient{
			HTTPRequester: c.HTTPRequester,
			Logger:        defaultLogger,
		}
	})
}

func ClientOptionWithDebugLogger(logger *zap.Logger) HTTPClientOption {
	return FuncHTTPClientOption(func(c *Client) {
		if logger == nil {
			logger = defaultLogger
		}
		c.HTTPRequester = &DebugClient{
			HTTPRequester: c.HTTPRequester,
			Logger:        logger,
		}
	})
}

func getClientTransport(c *Client) *http.Transport {
	if httpTransport, isHttpTransport := c.Transport.(*http.Transport); isHttpTransport {
		return httpTransport
	}
	panic("http client transport is not an http transport")
}

type HTTPRequester interface {
	// Do send an HTTP request defined in the r
	// Response can not be nil if err is nil
	Do(r *http.Request) (*http.Response, error)
}

type TokenClient struct {
	HTTPRequester

	Token string
}

func (tc *TokenClient) Do(r *http.Request) (*http.Response, error) {
	if tc.Token != "" {
		r.Header.Set(HeaderAuthorization, "Bearer "+tc.Token)
	}

	return tc.HTTPRequester.Do(r)
}

type HeaderClient struct {
	HTTPRequester

	Headers http.Header
}

func (hc *HeaderClient) Do(r *http.Request) (*http.Response, error) {
	if len(hc.Headers) == 0 {
		return hc.HTTPRequester.Do(r)
	}
	ExtendHeaders(r.Header, hc.Headers)
	return hc.HTTPRequester.Do(r)
}

type DebugClient struct {
	HTTPRequester

	Logger *zap.Logger
}

func (dc *DebugClient) Do(r *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	if dc.Logger != nil && dc.Logger.Level().Enabled(zap.DebugLevel) {
		defer NewHTTPRequestRecorder(dc.Logger, r, &resp, &err).Record()
	}
	resp, err = dc.HTTPRequester.Do(r)
	return resp, err
}
