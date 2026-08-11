package http

import (
	"context"
	"errors"
	"io"
	stdhttp "net/http"
	"net/netip"
	urlpkg "net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/datasources/internal"
	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/netx/dialerx"
	"github.com/duakc/lightddns/infra/netx/httpx"
	"github.com/duakc/lightddns/options"

	"github.com/itchyny/gojq"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestRequestContextHandleUsesJQForJSONResponse(t *testing.T) {
	rc := newTestRequestContext(t, stdhttp.MethodPost, "https://example.test/ip",
		matcher(t, ".ip", ""), requesterFunc(func(req *stdhttp.Request) (*stdhttp.Response, error) {
			require.Equal(t, stdhttp.MethodPost, req.Method)
			require.Equal(t, "unit-test", req.Header.Get("X-Test"))
			return testResponse(stdhttp.StatusOK, "application/json", `{"ip":"203.0.113.1"}`), nil
		}))
	rc.headers.Set("X-Test", "unit-test")

	got, err := rc.Handle(context.Background())
	require.NoError(t, err)
	require.Equal(t, parseAddrs("203.0.113.1"), got)
}

func TestRequestContextHandleDoesNotFallbackWhenJSONJQFails(t *testing.T) {
	rc := newTestRequestContext(t, stdhttp.MethodGet, "https://example.test/ip",
		matcher(t, ".ip", `fallback:\s+(\S+)`), requesterFunc(func(req *stdhttp.Request) (*stdhttp.Response, error) {
			return testResponse(stdhttp.StatusOK, "application/json", `{"ip":"not-an-ip","text":"fallback: 203.0.113.2"}`), nil
		}))

	_, err := rc.Handle(context.Background())
	require.Error(t, err)
}

func TestRequestContextHandleUsesRegexForTextResponse(t *testing.T) {
	rc := newTestRequestContext(t, stdhttp.MethodGet, "https://example.test/ip",
		matcher(t, "", `IP:\s+(\S+)`), requesterFunc(func(req *stdhttp.Request) (*stdhttp.Response, error) {
			return testResponse(stdhttp.StatusOK, "text/plain", "IP: 203.0.113.3\nIP: 2001:db8::3\n"), nil
		}))

	got, err := rc.Handle(context.Background())
	require.NoError(t, err)
	require.Equal(t, parseAddrs("203.0.113.3", "2001:db8::3"), got)
}

func TestRequestContextHandleFallsBackToPlainText(t *testing.T) {
	rc := newTestRequestContext(t, stdhttp.MethodGet, "https://example.test/ip",
		matcher(t, "", ""), requesterFunc(func(req *stdhttp.Request) (*stdhttp.Response, error) {
			return testResponse(stdhttp.StatusOK, "text/plain", "203.0.113.4\n2001:db8::4\n"), nil
		}))

	got, err := rc.Handle(context.Background())
	require.NoError(t, err)
	require.Equal(t, parseAddrs("203.0.113.4", "2001:db8::4"), got)
}

func TestRequestContextHandleCommonHTTPServiceResponses(t *testing.T) {
	tests := []struct {
		name        string
		rawURL      string
		jq          string
		regex       string
		contentType string
		body        string
		want        []netip.Addr
	}{
		{
			name:        "ipinfo.io JSON",
			rawURL:      "https://ipinfo.io",
			jq:          ".ip",
			contentType: "application/json",
			body:        `{"ip":"203.0.113.20","city":"example"}`,
			want:        parseAddrs("203.0.113.20"),
		},
		{
			name:        "api.ip.sb plain text",
			rawURL:      "https://api.ip.sb/ip",
			contentType: "text/plain",
			body:        "203.0.113.21\n",
			want:        parseAddrs("203.0.113.21"),
		},
		{
			name:        "api64.ipify.org JSON",
			rawURL:      "https://api64.ipify.org?format=json",
			jq:          ".ip",
			contentType: "application/json",
			body:        `{"ip":"2001:db8::22"}`,
			want:        parseAddrs("2001:db8::22"),
		},
		{
			name:        "myip.ipip.net text",
			rawURL:      "https://myip.ipip.net",
			regex:       `当前 IP：\s*(.+?)\s*来自于：`,
			contentType: "text/plain; charset=utf-8",
			body:        "当前 IP：203.0.113.23  来自于：中国 示例\n",
			want:        parseAddrs("203.0.113.23"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := newTestRequestContext(t, stdhttp.MethodGet, tt.rawURL,
				matcher(t, tt.jq, tt.regex), requesterFunc(func(req *stdhttp.Request) (*stdhttp.Response, error) {
					require.Equal(t, tt.rawURL, req.URL.String())
					return testResponse(stdhttp.StatusOK, tt.contentType, tt.body), nil
				}))

			got, err := rc.Handle(context.Background())
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRequestContextHandleRejectsBadStatus(t *testing.T) {
	rc := newTestRequestContext(t, stdhttp.MethodGet, "https://example.test/ip",
		matcher(t, "", ""), requesterFunc(func(req *stdhttp.Request) (*stdhttp.Response, error) {
			return testResponse(stdhttp.StatusTeapot, "text/plain", "203.0.113.5"), nil
		}))

	_, err := rc.Handle(context.Background())
	var badStatus *httpx.BadStatusCodeError
	require.ErrorAs(t, err, &badStatus)
	require.Equal(t, stdhttp.StatusTeapot, badStatus.Got)
}

func TestRequestContextHandleWrapsRequesterError(t *testing.T) {
	sentinel := errors.New("network down")
	rc := newTestRequestContext(t, stdhttp.MethodGet, "https://example.test/ip",
		matcher(t, "", ""), requesterFunc(func(req *stdhttp.Request) (*stdhttp.Response, error) {
			return nil, sentinel
		}))

	_, err := rc.Handle(context.Background())
	var responseErr *httpx.BaseResponseError
	require.ErrorAs(t, err, &responseErr)
	require.ErrorIs(t, err, sentinel)
}

func TestNewHTTPDatasourceBuildsIPv4RequestWithDefaultsAndDebug(t *testing.T) {
	ds, err := New(context.Background(), zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel)), options.HTTPDatasourceOption{
		URL: badyaml.URL{
			URL: mustURL(t, "http://192.0.2.6/ip"),
			Raw: "http://192.0.2.6/ip",
		},
		Connect: options.ConnectOption{
			DialStrategy: dialerx.DialOnlyIPv4,
		},
		HTTP: options.HTTPOption{
			HTTPDebug: true,
		},
		Headers: badyaml.HTTPHeader{
			Header: stdhttp.Header{"X-Test": []string{"configured"}},
		},
	})
	require.NoError(t, err)

	httpds := ds.(*Httpds)
	require.NotNil(t, httpds.v4)
	require.Nil(t, httpds.v6)
	require.Equal(t, constant.HTTPUserAgent, httpds.v4.headers.Get("User-Agent"))
	require.Equal(t, "configured", httpds.v4.headers.Get("X-Test"))

	client := httpds.v4.requester.(*httpx.Client)
	_, debugEnabled := client.HTTPRequester.(*httpx.DebugClient)
	require.True(t, debugEnabled)
}

func TestNewHTTPDatasourceHonorsIPv6OnlyDialStrategy(t *testing.T) {
	ds, err := New(context.Background(), zaptest.NewLogger(t), options.HTTPDatasourceOption{
		URL: badyaml.URL{
			URL: mustURL(t, "https://example.test/ip"),
			Raw: "https://example.test/ip",
		},
		Connect: options.ConnectOption{
			DialStrategy: dialerx.DialOnlyIPv6,
		},
	})
	require.NoError(t, err)

	httpds := ds.(*Httpds)
	require.Nil(t, httpds.v4)
	require.NotNil(t, httpds.v6)
}

type requesterFunc func(*stdhttp.Request) (*stdhttp.Response, error)

func (f requesterFunc) Do(req *stdhttp.Request) (*stdhttp.Response, error) {
	return f(req)
}

func newTestRequestContext(
	t *testing.T,
	method string,
	rawURL string,
	matcher *internal.IPMatcher,
	requester httpx.HTTPRequester,
) *requestContext {
	t.Helper()

	return &requestContext{
		method:    method,
		url:       mustURL(t, rawURL),
		headers:   stdhttp.Header{},
		matcher:   matcher,
		requester: requester,
	}
}

func matcher(t *testing.T, jqExpr, regexExpr string) *internal.IPMatcher {
	t.Helper()

	var jq *gojq.Query
	if jqExpr != "" {
		var err error
		jq, err = gojq.Parse(jqExpr)
		require.NoError(t, err)
	}

	var re *regexp.Regexp
	if regexExpr != "" {
		re = regexp.MustCompile(regexExpr)
	}

	return internal.NewDefaultIPMatcher(jq, re)
}

func mustURL(t *testing.T, raw string) *urlpkg.URL {
	t.Helper()

	parsed, err := urlpkg.Parse(raw)
	require.NoError(t, err)
	return parsed
}

func testResponse(statusCode int, contentType string, body string) *stdhttp.Response {
	resp := &stdhttp.Response{
		StatusCode: statusCode,
		Header:     stdhttp.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	if contentType != "" {
		resp.Header.Set("Content-Type", contentType)
	}
	return resp
}

func parseAddrs(values ...string) []netip.Addr {
	addrs := make([]netip.Addr, 0, len(values))
	for _, value := range values {
		addrs = append(addrs, netip.MustParseAddr(value))
	}
	return addrs
}
