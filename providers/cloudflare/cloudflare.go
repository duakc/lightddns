package cloudflare

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/ctxservice"
	"github.com/duakc/lightddns/infra/generic"
	"github.com/duakc/lightddns/infra/netxx"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	providerpkg "github.com/duakc/lightddns/providers"
	"github.com/duakc/lightddns/providers/cloudflare/internal"
	"go.uber.org/zap"
)

func init() {
	adapter.Register(
		adapter.ProviderRegister,
		constpkg.ProviderTypeCloudflare,
		New,
	)
}

func New(ctx context.Context, option options.OptionProviderCloudflare) (adapter.Provider, error) {
	upstreamLogger := ctxservice.Lookup[*zap.Logger](ctx, zaplog.LoggerKey{})
	logger := zaplog.ExtendName(upstreamLogger, option.Name)
	if option.Token == "" {
		return nil, fmt.Errorf("cloudflare(%s): %w", option.Name, providerpkg.ErrRequireToken)
	}

	cf := &Cloudflare{
		logger:  logger,
		client:  internal.NewClient(ctx, option.Token),
		zones:   new(generic.SyncMap[string, string]),
		name:    option.Name,
		proxied: option.Proxy,
	}

	return cf, nil
}

type Cloudflare struct {
	logger *zap.Logger
	client *internal.Client

	name      string
	zones     *generic.SyncMap[string, string]
	zoneMutex sync.Mutex

	proxied      bool
	privateRoute bool
}

func (c *Cloudflare) Type() string {
	return constpkg.ProviderTypeCloudflare
}

func (c *Cloudflare) Name() string {
	return c.name
}

func (c *Cloudflare) getZoneID(ctx context.Context, domain string) (string, error) {
	// TODO: add domain suffix match
	if existedZoneID, existed := c.zones.Load(domain); existed {
		return existedZoneID, nil
	}
	c.zoneMutex.Lock()
	defer c.zoneMutex.Unlock()
	if existedZoneID, existed := c.zones.Load(domain); existed {
		return existedZoneID, nil
	}
	return c.updateZoneID(ctx, domain)
}

func (c *Cloudflare) updateZoneID(ctx context.Context, domain string) (string, error) {
	zoneName := c.client.ListZoneName(domain)
	for page, err := zoneName.Next(ctx); err != io.EOF; page, err = zoneName.Next(ctx) {
		if err != nil {
			return "", err
		}
		for i := 0; i < len(page); i++ {
			zone := page[i]
			if len(zone.ID) < 4 {
				return "",
					fmt.Errorf("zone id too short for %s", domain)
			}
			if !netxx.IsSubDomain(domain, zone.Name) {
				continue
			}

			c.logger.Info("found zone id",
				zap.String("domain", domain))

			c.zones.Store(zone.Name, zone.ID)
			c.zones.Store(domain, zone.ID)
			return zone.ID, nil
		}
	}
	return "", fmt.Errorf("not found zone id for %s", domain)
}
