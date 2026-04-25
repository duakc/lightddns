package httpds

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"regexp"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	datasourcepkg "github.com/duakc/lightddns/datasources"
	"github.com/duakc/lightddns/infra/httpxx"
	"github.com/duakc/lightddns/infra/lookctx"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/resolvectl"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"

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

func New(ctx context.Context, option options.HTTPDatasourceOption) (adapter.Datasource, error) {
	if option.Method == "" {
		option.Method = http.MethodGet
	}
	isDomainName := true
	if !netool.IsDomainName(option.Url.URL.Host) {
		ipAddress := option.Url.URL.Host
		addr, err := netip.ParseAddr(ipAddress)
		if err != nil {
			return nil, fmt.Errorf("unknown address: %w: %w", err, dialerx.ErrNoAddressToDialer)
		}
		isDomainName = false
		if netool.IsIPv4(addr) {
			option.DialStrategy = dialerx.DialOnlyIPv4
		} else {
			option.DialStrategy = dialerx.DialOnlyIPv6
		}
	}
	dialerOptions, err := option.ConnectOption.Options()
	if err != nil {
		return nil, fmt.Errorf("dialer: %w", err)
	}
	httpOptions, err := option.HTTPOption.Options()
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}

	var (
		logger = datasourcepkg.NewLogger(lookctx.LookupPtr[zap.Logger](ctx), option.AbstractDatasourceOption)
		v4, v6 *requestContext
	)

	connectDialer := dialerx.NewDialerWithOption(dialerOptions...)
	if isDomainName {
		connectDialer = resolvectl.NewDialer(ctx, connectDialer, mt.Must(option.DNS.NewTransport(ctx, connectDialer)))
	}

	if option.DialStrategy != dialerx.DialOnlyIPv6 {
		// enable ipv4
		httpClient := httpxx.NewClient(append(httpOptions,
			httpxx.ClientOptionWithDialer(&dialerx.NetworkDialer{Network: "tcp4", Dialer: connectDialer}))...)
		v4, err = newRequestContext(string(option.Method), option.Url.Raw,
			option.Headers.Header, httpClient,
			cmp.Or(option.MatchJson.Str, option.MatchJson.Obj.V4),
			cmp.Or(option.MatchRegex.Str, option.MatchRegex.Obj.V4))
		if err != nil {
			return nil, err
		}
	}
	if option.DialStrategy != dialerx.DialOnlyIPv4 {
		// enable ipv6
		httpClient := httpxx.NewClient(append(httpOptions,
			httpxx.ClientOptionWithDialer(&dialerx.NetworkDialer{Network: "tcp6", Dialer: connectDialer}))...)
		v6, err = newRequestContext(string(option.Method), option.Url.Raw,
			option.Headers.Header, httpClient,
			cmp.Or(option.MatchJson.Str, option.MatchJson.Obj.V6),
			cmp.Or(option.MatchRegex.Str, option.MatchRegex.Obj.V6))
		if err != nil {
			return nil, err
		}
	}
	httpds := &Httpds{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),
		logger:              logger,
	}
	httpds.v4 = v4
	httpds.v6 = v6
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
	logger := c.logger

	logger.Debug("ipv4 request")
	return c.v4.Handle(ctx)
}

func (c *Httpds) IPv6(ctx context.Context) ([]netip.Addr, error) {
	if c.v6 == nil {
		return []netip.Addr{}, nil
	}
	logger := c.logger

	logger.Debug("ipv6 request")
	return c.v6.Handle(ctx)
}

func (c *Httpds) IP(ctx context.Context) ([]netip.Addr, error) {
	return adapter.MergeDualStackDatasourceIP(ctx, c)
}

type requestContext struct {
	method  string
	url     string
	headers http.Header

	requester httpxx.HTTPRequester

	jsonMatch  *gojq.Query
	regexMatch *regexp.Regexp
}

func newRequestContext(method string, url string, headers http.Header,
	requester httpxx.HTTPRequester, jq string, re string,
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
	R.method = method
	R.url = url
	R.headers = headers
	R.requester = requester

	return R, nil
}

func (rc *requestContext) Handle(ctx context.Context) (addresses []netip.Addr, err error) {
	R := httpxx.NewReqConfig(rc.method, rc.url)
	R.ExtendHeader = rc.headers
	request, err := R.ToRequestContext(ctx)
	if err != nil {
		return nil, err
	}
	response, err := rc.requester.Do(request)
	if err != nil {
		return nil, httpxx.NewBaseResponseError(err, R.Method, "get ip from remote")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, &httpxx.BadStatusCodeError{Got: response.StatusCode}
	}
	const maxBodySize = 10 * 1024 * 1024

	if response.Header.Get("Content-Type") == "application/json" && rc.jsonMatch != nil {
		var jsonObject any
		if err := json.NewDecoder(io.LimitReader(response.Body, maxBodySize)).Decode(&jsonObject); err != nil {
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
		buffer, err := io.ReadAll(io.LimitReader(response.Body, maxBodySize))
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
