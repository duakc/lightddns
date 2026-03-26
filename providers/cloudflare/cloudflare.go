package cloudflare

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/duakc/lightddns/adapter"
	CST "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/common/generic"
	"github.com/duakc/lightddns/infra/ctxservice"
	"github.com/duakc/lightddns/infra/netxx"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/providers/cloudflare/internal"
	"go.uber.org/zap"
)

func New(ctx context.Context, option options.OptionProviderCloudflare) (adapter.Provider, error) {
	logger := ctxservice.Lookup[*zap.Logger](ctx, zaplog.LoggerKey{})

}

type cloudflare struct {
	logger *zap.Logger
	client *internal.Client

	name      string
	zones     *generic.SyncMap[string, string]
	zoneMutex sync.Mutex

	proxied      bool
	privateRoute bool
}

func (c *cloudflare) Type() string {
	return CST.ProviderTypeCloudflare
}

func (c *cloudflare) Name() string {
	return c.name
}

func (c *cloudflare) getZoneID(ctx context.Context, domain string) (string, error) {
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

func (c *cloudflare) updateZoneID(ctx context.Context, domain string) (string, error) {
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
