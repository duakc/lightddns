package cloudflare

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/generic"
	"github.com/duakc/lightddns/infra/httpxx"
	"github.com/duakc/lightddns/infra/lookctx"
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

func New(ctx context.Context, option options.CloudflareProviderOption) (adapter.Provider, error) {
	if option.Token == "" {
		return nil, fmt.Errorf("cloudflare(%s): %w", option.Name, providerpkg.ErrRequireToken)
	}
	dialerOptions, err := option.ConnectOption.Options()
	if err != nil {
		return nil, err
	}
	clientOptions, err := option.HTTPOption.Options()
	if err != nil {
		return nil, err
	}
	clientOptions = append(clientOptions,
		httpxx.ClientOptionWithToken(option.Token),
		httpxx.ClientOptionWithDialer(
			netxx.NewDialerWithOption(dialerOptions...)))

	cf := &Cloudflare{
		logger: providerpkg.NewLogger(
			lookctx.Lookup[zaplog.LoggerKey, *zap.Logger](ctx),
			option.AbstractProviderOption,
		),
		client: internal.NewClient(ctx, httpxx.NewClient(clientOptions...)),
		zones:  new(generic.SyncMap[string, string]),

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
	var zoneID string

	logger.Info("search zone id from upstream", zap.String("domain", domain))

	for page, err := zoneName.Next(ctx); err != io.EOF; page, err = zoneName.Next(ctx) {
		if err != nil {
			return "", err
		}
		for i := 0; i < len(page); i++ {
			zone := page[i]
			if !netxx.IsDomainName(zone.Name) {
				logger.Warn("upstream return a bad domain",
					zap.String("domain", domain),
					zap.String("zone_name", zone.Name))

				continue
			}

			logger.Info("found zone id",
				zap.String("domain", domain),
				zap.String("zone_name", zone.Name))
			if zone.Name == domain {
				zoneID = zone.ID
			}

			c.zones.Store(zone.Name, zone.ID)
			c.zones.Store(domain, zone.ID)
		}
	}
	if zoneID != "" {
		return zoneID, nil
	} else if zoneID = c.fullMatchDomainZoneID(domain); zoneID != "" {
		return zoneID, nil
	}
	return "", fmt.Errorf("zone id for %s not found", domain)
}

func (c *Cloudflare) fullMatchDomainZoneID(domain string) string {
	domainToken := strings.Split(domain, ".")
	for i := len(domainToken) - 2; i > -1; i-- {
		fullDomain := strings.Join(domainToken[i:], ".")
		if existedZoneID, existed := c.zones.Load(fullDomain); existed {
			return existedZoneID
		}
	}
	return ""
}
