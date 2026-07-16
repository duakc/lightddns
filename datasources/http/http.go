package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	urlpkg "net/url"
	"regexp"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/options/castoption"

	"github.com/itchyny/gojq"
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

	v4URL, v6URL := option.URL.IPv4, option.URL.IPv6
	if v4URL.Raw == "" && v6URL.Raw == "" {
		return nil, fmt.Errorf("url: at least one of ipv4 or ipv6 must be specified")
	}

	// If a URL host is a literal IP, force its stack and disable the other.
	for stack, url := range map[string]*badyaml.URL{"ipv4": &v4URL, "ipv6": &v6URL} {
		if url.Raw == "" || domains.IsDomainName(url.URL.Host) {
			continue
		}
		addr, err := netip.ParseAddr(url.URL.Host)
		if err != nil {
			return nil, fmt.Errorf("url.%s: unknown address: %w: %w", stack, err, dialerx.ErrNoAddressToDialer)
		}
		if (stack == "ipv4") != netool.IsIPv4(addr) {
			return nil, fmt.Errorf("url.%s: address family mismatch: %s", stack, addr)
		}
	}

	needDNS := (v4URL.Raw != "" && domains.IsDomainName(v4URL.URL.Host)) ||
		(v6URL.Raw != "" && domains.IsDomainName(v6URL.URL.Host))

	var v4, v6 *requestContext

	connectDialer, err := castoption.BuildDialer(option.Connect)
	if err != nil {
		return nil, err
	}
	if needDNS {
		resolveDialer, err := castoption.BuildResolveDialer(logger, connectDialer, option.DNS)
		if err != nil {
			return nil, err
		}
		connectDialer = resolveDialer
	}

	httpOptions, err := castoption.HTTPOptionToHTTPXOptions(option.HTTP)
	if err != nil {
		return nil, fmt.Errorf("building http options: %w", err)
	}

	if v4URL.Raw != "" && option.Connect.DialStrategy != dialerx.DialOnlyIPv6 {
		httpClient := httpx.NewClient(append(httpOptions,
			httpx.ClientOptionWithDialer(&dialerx.NetworkDialer{Network: "tcp4", Dialer: connectDialer}))...)
		v4, err = newRequestContext(string(option.Method), v4URL.Raw,
			option.Headers.Header, httpClient,
			option.JSON.IPv4, option.Regex.IPv4)
		if err != nil {
			return nil, err
		}
	}

	if v6URL.Raw != "" && option.Connect.DialStrategy != dialerx.DialOnlyIPv6 {
		httpClient := httpx.NewClient(append(httpOptions,
			httpx.ClientOptionWithDialer(&dialerx.NetworkDialer{Network: "tcp6", Dialer: connectDialer}))...)
		v6, err = newRequestContext(string(option.Method), v6URL.Raw,
			option.Headers.Header, httpClient,
			option.JSON.IPv6, option.Regex.IPv6)
		if err != nil {
			return nil, err
		}
	}

	httpds := &Httpds{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),

		logger: logger,
		v4:     v4,
		v6:     v6,
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
	return adapter.MergeDualStackDatasourceIP(ctx, c)
}

type requestContext struct {
	method  string
	url     *urlpkg.URL
	headers http.Header

	requester httpx.HTTPRequester

	jsonMatch  *gojq.Query
	regexMatch *regexp.Regexp
}

func newRequestContext(method string, url string, headers http.Header,
	requester httpx.HTTPRequester, jq string, re string,
) (*requestContext, error) {
	R := new(requestContext)
	var err error
	if jq != "" {
		if R.jsonMatch, err = gojq.Parse(jq); err != nil {
			return nil, fmt.Errorf("MatchJson: %w", err)
		}
	}
	if re != "" {
		if R.regexMatch, err = regexp.Compile(re); err != nil {
			return nil, fmt.Errorf("MatchRegex: %w", err)
		}
	}
	parsedURL, err := urlpkg.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	R.method = method
	R.url = parsedURL
	R.headers = headers
	R.requester = requester

	return R, nil
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

	limitedReader := io.LimitReader(response.Body, constpkg.HTTPMaxBodySize)

	if httpx.IsJsonContentType(response.Header.Get("Content-Type")) && rc.jsonMatch != nil {
		var jsonObject any
		if err := json.NewDecoder(limitedReader).Decode(&jsonObject); err != nil {
			return nil, fmt.Errorf("decode JSON: %w", err)
		}
		iter := rc.jsonMatch.RunWithContext(ctx, jsonObject)
		for val, ok := iter.Next(); ok; val, ok = iter.Next() {
			switch x := val.(type) {
			case error:
				if haltErr, ok := errors.AsType[*gojq.HaltError](x); ok && haltErr.Value() == nil {
					break
				}
				return nil, fmt.Errorf("jq execution error: %w", x)
			case string:
				addr, err := netip.ParseAddr(x)
				if err != nil {
					return nil, err
				}
				addresses = append(addresses, addr)
			}
		}
	} else {
		buffer, err := io.ReadAll(limitedReader)
		if err != nil {
			return nil, fmt.Errorf("read response.Body: %w", err)
		}

		if rc.regexMatch != nil {
			for _, x := range rc.regexMatch.FindAllSubmatch(buffer, -1) {
				if len(x) > 1 {
					xx := x[1]
					addr, err := netip.ParseAddr(string(xx))
					if err != nil {
						return nil, err
					}
					addresses = append(addresses, addr)
				}
			}
		} else {
			addr, err := netip.ParseAddr(string(bytes.TrimSpace(buffer)))
			if err != nil {
				return nil, err
			}
			addresses = append(addresses, addr)
		}
	}
	return addresses, nil
}
