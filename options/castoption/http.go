package castoption

import (
	"fmt"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/netx/dialerx"
	"github.com/duakc/lightddns/infra/netx/httpx"
	"github.com/duakc/lightddns/options"

	"go.uber.org/zap"
)

func BuildHTTPClient(
	rawOption options.HTTPOption,
	underlay dialerx.Dialer,
	logger *zap.Logger,
) (httpx.HTTPRequester, error) {
	httpClientOptions, err := HTTPOptionToHTTPXOptions(rawOption)
	if err != nil {
		return nil, fmt.Errorf("build http client option: %w", err)
	}
	if underlay == nil {
		underlay = dialerx.NewDialerWithOption()
	}

	// Add User-Agent
	httpClientOptions = append(httpClientOptions, httpx.ClientOptionWithHeader(
		httpx.HeaderUserAgent, constpkg.HTTPUserAgent))

	if rawOption.HTTPDebug && logger.Level().Enabled(zap.DebugLevel) {
		httpClientOptions = append(httpClientOptions, httpx.ClientOptionWithDebugLogger(logger))
	}

	return httpx.NewClient(underlay, httpClientOptions...), nil
}

func HTTPOptionToHTTPXOptions(rawOption options.HTTPOption) ([]httpx.HTTPClientOption, error) {
	var httpClientOptions []httpx.HTTPClientOption

	if rawOption.HTTPProxy != "" || rawOption.HTTPSProxy != "" {
		if rawOption.HTTPProxy == "" {
			rawOption.HTTPProxy = rawOption.HTTPSProxy
		}
		if rawOption.HTTPSProxy == "" {
			rawOption.HTTPSProxy = rawOption.HTTPProxy
		}
		if rawOption.HTTPProxy == "" && rawOption.HTTPSProxy == "" && rawOption.UseSystemProxy {
			httpClientOptions = append(httpClientOptions, httpx.ClientOptionEnableProxy())
		} else {
			httpClientOptions = append(httpClientOptions, httpx.ClientOptionWithProxy(rawOption.HTTPProxy, rawOption.HTTPSProxy, ""))
		}
	}
	return httpClientOptions, nil
}
