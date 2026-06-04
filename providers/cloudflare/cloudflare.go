package cloudflare

import (
	"context"
	"fmt"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/ddnsmetric"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"
	"github.com/duakc/lightddns/infra/netool/resolvectl"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/providers/cloudflare/internal"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

const ProviderType = constpkg.ProviderTypeCloudflare

func init() {
	adapter.Register(
		adapter.ProviderRegister,
		ProviderType,
		New,
	)
}

func New(ctx context.Context, logger *zap.Logger, option options.CloudflareProviderOption) (adapter.Provider, error) {
	if option.Token == "" {
		return nil, fmt.Errorf("cloudflare(%s): token is empty", option.Name)
	}

	dialerOptions, err := option.ConnectOption.Options()
	if err != nil {
		return nil, err
	}
	clientOptions, err := option.HTTPOption.Options()
	if err != nil {
		return nil, err
	}

	connectDialer := dialerx.NewDialerWithOption(dialerOptions...)
	clientOptions = append(clientOptions,
		httpx.ClientOptionWithDialer(
			resolvectl.NewDialer(connectDialer,
				mt.Must(option.DNS.NewTransport(ctx, connectDialer)), resolvectl.DefaultResolveClient)))

	return &Cloudflare{
		AbstractManagedType: adapter.NewManagedType(ProviderType, option.Name),
		client: internal.NewClient(logger, option.Name,
			httpx.NewClient(clientOptions...), option.Token),

		proxied: option.Proxy,
		logger:  logger,
	}, nil
}

type Cloudflare struct {
	adapter.AbstractManagedType

	logger *zap.Logger
	client *internal.Client

	zones ddnsx.DomainIdCache

	proxied      bool
	privateRoute bool
}

func (c *Cloudflare) RegisterMetrics(factory ddnsmetric.Factory) {
	c.client.RegisterMetrics(factory)
}

func (c *Cloudflare) Close() error { return nil }

func (c *Cloudflare) getZoneID(ctx context.Context, domain string) (string, error) {
	if !domains.IsDomainName(domain) {
		return "", fmt.Errorf("bad domain name: %s", domain)
	}
	zoneID := c.zones.LoadOrStore(ctx, domain, c.client)
	if zoneID == "" {
		return "", fmt.Errorf("zone id for %s not found", domain)
	}
	return zoneID, nil
}
