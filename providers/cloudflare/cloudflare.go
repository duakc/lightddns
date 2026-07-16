package cloudflare

import (
	"context"
	"fmt"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/options/castoption"
	"github.com/duakc/lightddns/providers/cloudflare/internal"

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
		return nil, fmt.Errorf("token is empty")
	}

	_, _, httpClient, err := castoption.BuildHTTPClientFromScratch(
		logger, option.Connect, option.DNS, option.HTTP)
	if err != nil {
		return nil, err
	}

	return &Cloudflare{
		AbstractManagedType: adapter.NewManagedType(ProviderType, option.Name),
		client: internal.NewClient(ctx, logger, option.Name,
			httpClient, option.Token),

		proxied: option.Proxy,
		logger:  logger.Named("client"),
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
