package castoption

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/duakc/lightddns/infra/netx/dialerx"
	"github.com/duakc/lightddns/infra/netx/resolvectl"
	"github.com/duakc/lightddns/infra/netx/resolvectl/transports"
	"github.com/duakc/lightddns/options"

	"go.uber.org/zap"
)

func BuildResolveDialer(DNS options.DNSOption, underlay dialerx.Dialer, logger *zap.Logger) (dialerx.Dialer, error) {
	if DNS.Type == "" {
		DNS.Type = transports.TransportTypeSystem
	}
	transportOptions := DNSOptionToTransportsTransportOptions(logger, underlay, DNS)
	dnsTransport, err := transports.NewTransport(transportOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create dns transport: %s", err)
	}

	var resolveClient resolvectl.ResolveClient
	if DNS.Type != transports.TransportTypeSystem {
		// use independent cache with a lower cache size.
		resolveClient = resolvectl.NewResolverWithCache(logger,
			resolvectl.NewResolverCacheSize(1024))
	}

	if resolveClient == nil {
		resolveClient = resolvectl.NewResolver(logger)
	}

	return resolvectl.NewDialer(underlay, dnsTransport, resolveClient), nil
}

func DNSOptionToTransportsTransportOptions(logger *zap.Logger,
	dialer dialerx.Dialer,
	rawOption options.DNSOption,
) transports.TransportOptions {
	return transports.TransportOptions{
		Logger:     logger,
		Type:       strings.ToLower(rawOption.Type),
		Dialer:     dialer,
		Server:     rawOption.Server,
		ServerPort: rawOption.Port,
		TLSConfig:  &tls.Config{},
	}
}
