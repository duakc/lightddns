package cloudflare

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/adapter/providerx"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/options/castoption"

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

var _ adapter.Provider = (*Cloudflare)(nil)

type Cloudflare struct {
	adapter.AbstractManagedType

	reconciler *ddnsx.Reconciler[DNSRecord]
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

	client := NewClient(logger, httpClient, option.Token,
		option.Proxy, false, constpkg.ThisRecordIsManagedByLightddns)
	observed := providerx.NewMetricsClientFromContext(
		ctx, option.Name, ProviderType, client,
	)
	return &Cloudflare{
		AbstractManagedType: adapter.NewManagedType(ProviderType, option.Name),
		reconciler:          ddnsx.NewReconciler(logger.Named("client"), observed),
	}, nil
}

func (c *Cloudflare) Close() error { return nil }

func (c *Cloudflare) Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error) {
	return c.reconciler.Diff(ctx, domain, addr)
}

func (c *Cloudflare) Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (bool, error) {
	return c.reconciler.Update(ctx, domain, ttl, addr)
}
