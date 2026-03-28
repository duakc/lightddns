package cloudflare

import (
	"context"
	"fmt"
	"io"
	"strings"
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
	if !netxx.IsDomainName(domain) {
		return "", fmt.Errorf("bad domain name")
	}
	if existedZoneID := c.fullMatchDomainZoneID(domain); existedZoneID != "" {
		return existedZoneID, nil
	}

	c.zoneMutex.Lock()
	defer c.zoneMutex.Unlock()
	if existedZoneID := c.fullMatchDomainZoneID(domain); existedZoneID != "" {
		return existedZoneID, nil
	}
	return c.updateZoneID(ctx, domain)
}

func (c *Cloudflare) updateZoneID(ctx context.Context, domain string) (string, error) {
	logger := c.logger
	zoneName := c.client.ListZones()
	zoneID := ""
	for page, err := zoneName.Next(ctx); err != io.EOF; page, err = zoneName.Next(ctx) {
		if err != nil {
			return "", err
		}
		for i := 0; i < len(page); i++ {
			zone := page[i]
			if !netxx.IsDomainName(zone.Name) {
				logger.Warn("upstream return a wrong domain",
					zap.String("domain", zone.Name))
				continue
			}
			if netxx.IsSubDomain(domain, zone.Name) {
				zoneID = zone.ID
			}

			logger.Info("found zone id", zap.String("domain", domain))

			c.zones.Store(zone.Name, zone.ID)
			c.zones.Store(domain, zone.ID)
		}
	}
	if zoneID != "" {
		return zoneID, nil
	}
	return "", fmt.Errorf("not found zone id for %s", domain)
}

func (c *Cloudflare) fullMatchDomainZoneID(domain string) string {
	domainToken := strings.Split(domain, ".")
	for i := len(domainToken) - 1; i > -1; i-- {
		fullDomain := strings.Join(domainToken[i:], ".")
		if existedZoneID, existed := c.zones.Load(fullDomain); existed {
			return existedZoneID
		}
	}
	return ""
}
