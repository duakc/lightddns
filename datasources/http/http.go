package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/netip"
	urlpkg "net/url"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/datasourcex"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/datasources/internal"
	"github.com/duakc/lightddns/infra/netx/dialerx"
	"github.com/duakc/lightddns/infra/netx/domains"
	"github.com/duakc/lightddns/infra/netx/httpx"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/options/castoption"

	"github.com/duakc/mt/freebuf"

	"go.uber.org/zap"
)

const DatasourceType = constpkg.DatasourceTypeHTTP

func init() {
	adapter.Register(
		adapter.DatasourceRegister,
		DatasourceType,
		New,
	)
}

func New(ctx context.Context, logger *zap.Logger, option options.HTTPDatasourceOption) (adapter.Datasource, error) {
	if option.Method == "" {
		option.Method = http.MethodGet
	}

	needDNS := option.URL.Raw != "" && domains.IsDomainName(option.URL.URL.Host)

	connectDialer, err := castoption.BuildDialer(option.Connect)
	if err != nil {
		return nil, err
	}

	if needDNS {
		resolveDialer, err := castoption.BuildResolveDialer(option.DNS, connectDialer, logger)
		if err != nil {
			return nil, err
		}
		connectDialer = resolveDialer
	}

	httpOptions, err := castoption.HTTPOptionToHTTPXOptions(option.HTTP)
	if err != nil {
		return nil, fmt.Errorf("building http options: %w", err)
	}

	customHeaders := maps.Clone(option.Headers.Header)

	if len(customHeaders) == 0 {
		customHeaders = http.Header{}
	}

	if customHeaders.Get("User-Agent") == "" {
		customHeaders.Set("User-Agent", constpkg.HTTPUserAgent)
	}

	// the custom header must have on element (User-Agent)
	httpOptions = append(httpOptions, httpx.ClientOptionWithHeaders(customHeaders))

	requests := &requestContext{
		method:  string(option.Method),
		url:     option.URL.URL,
		headers: customHeaders,
		matcher: internal.NewDefaultIPMatcher(
			option.Match.JQ.Cast(),
			option.Match.Regexp.Cast(),
		),
	}

	var v4Request, v6Request requestContext
	if option.Connect.DialStrategy != dialerx.DialOnlyIPv4 {
		v6Request = *requests
		v6Request.requester = httpx.NewClient(&dialerx.NetworkDialer{Network: "tcp6", Dialer: connectDialer},
			httpOptions...)
	}
	if option.Connect.DialStrategy != dialerx.DialOnlyIPv6 {
		v4Request = *requests
		v4Request.requester = httpx.NewClient(&dialerx.NetworkDialer{Network: "tcp4", Dialer: connectDialer},
			httpOptions...)
	}

	httpds := &Httpds{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),

		logger: logger,
		v4:     &v4Request,
		v6:     &v6Request,
	}
	return httpds, nil
}

type Httpds struct {
	adapter.AbstractManagedType

	logger *zap.Logger
	v4     *requestContext
	v6     *requestContext
}

func (c *Httpds) IPv4(ctx context.Context) ([]netip.Addr, error) {
	if c.v4 == nil {
		return []netip.Addr{}, nil
	}
	return c.v4.Handle(ctx)
}

func (c *Httpds) IPv6(ctx context.Context) ([]netip.Addr, error) {
	if c.v6 == nil {
		return []netip.Addr{}, nil
	}
	return c.v6.Handle(ctx)
}

func (c *Httpds) IP(ctx context.Context) ([]netip.Addr, error) {
	return datasourcex.MergeDualStackDatasourceIP(ctx, c)
}

type requestContext struct {
	method  string
	url     *urlpkg.URL
	headers http.Header

	requester httpx.HTTPRequester

	matcher *internal.IPMatcher
}

func (rc *requestContext) Handle(ctx context.Context) (addresses []netip.Addr, err error) {
	R := httpx.NewReqConfig(rc.method, rc.url)
	R.ExtendHeader = rc.headers
	request, err := R.ToRequestContext(ctx)
	if err != nil {
		return nil, err
	}
	response, err := rc.requester.Do(request)
	if err != nil {
		return nil, httpx.NewBaseResponseError(err, R.Method, "get ip from remote")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, &httpx.BadStatusCodeError{Got: response.StatusCode}
	}

	buffer := freebuf.NewSerialLimited(constpkg.HTTPMaxBodySize)
	defer buffer.FreeMe()

	if _, readErr := buffer.ReadFrom(response.Body); readErr != nil && !errors.Is(readErr, io.ErrShortBuffer) {
		return nil, readErr
	}

	// this differs from `IPMatcher.Try`.
	// HTTP can carry information to determine whether the content is JSON or plain text.
	//
	// therefore, we first check if the returned value is JSON.
	// if it is JSON and jQuery failed, we simply return an error;
	// proceeding with the subsequent logic is pointless.
	if httpx.IsJsonContentType(response.Header.Get("Content-Type")) && rc.matcher.JQ != nil {
		addresses, err = rc.matcher.JSON(ctx, buffer.Bytes())
		if err != nil {
			return nil, err
		}
	}

	if rc.matcher.Regexp != nil {
		addresses, err = rc.matcher.Re(buffer.Bytes())
		if err == nil && len(addresses) > 0 {
			return addresses, nil
		}
	}

	return rc.matcher.Plain(buffer.Bytes())
}
