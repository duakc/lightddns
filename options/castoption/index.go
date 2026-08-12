package castoption

import (
	"fmt"

	"github.com/duakc/lightddns/infra/netx/dialerx"
	"github.com/duakc/lightddns/infra/netx/httpx"
	"github.com/duakc/lightddns/options"

	"go.uber.org/zap"
)

func BuildHTTPClientFromScratch(
	logger *zap.Logger,
	connectOption options.ConnectOption,
	dnsOption options.DNSOption,
	httpOption options.HTTPOption,
) (rawDialer, resolveDialer dialerx.Dialer, httpClient httpx.HTTPRequester, err error) {
	rawDialer, err = BuildDialer(connectOption)
	if err != nil {
		err = fmt.Errorf("building raw dialer: %w", err)
		return
	}

	httpDialer := rawDialer
	if dnsOption.Enabled {
		resolveDialer, err = BuildResolveDialer(dnsOption, rawDialer, logger)
		if err != nil {
			err = fmt.Errorf("building resolve dialer: %w", err)
			return
		}
		httpDialer = resolveDialer
	} else {
		resolveDialer = rawDialer
	}
	httpClient, err = BuildHTTPClient(httpOption, httpDialer, logger)
	if err != nil {
		err = fmt.Errorf("building http client: %w", err)
		return
	}
	return rawDialer, resolveDialer, httpClient, nil
}
