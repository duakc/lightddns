package castoption

import (
	"fmt"

	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/httpx"
	"github.com/duakc/lightddns/options"
)

func BuildHTTPClient(underlay dialerx.Dialer, rawOption options.HTTPOption) (httpx.HTTPRequester, error) {
	httpClientOptions, err := HTTPOptionToHTTPXOptions(rawOption)
	if err != nil {
		return nil, fmt.Errorf("build http client option: %w", err)
	}
	if underlay == nil {
		underlay = dialerx.NewDialerWithOption()
	}
	httpClientOptions = append(httpClientOptions, httpx.ClientOptionWithDialer(underlay))
	return httpx.NewClient(httpClientOptions...), nil
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
