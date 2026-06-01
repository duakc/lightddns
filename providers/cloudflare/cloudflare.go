package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/ddnsmetric"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/metrics"
	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"
	"github.com/duakc/lightddns/infra/netool/resolvectl"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/providers/cloudflare/internal"

	"github.com/duakc/mt"
	"github.com/duakc/mt/common/generic"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

const (
	metricRequestTotal           = constpkg.MetricProviderRequestTotal
	metricRequestFailureTotal    = constpkg.MetricProviderRequestFailureTotal
	metricRequestDurationSeconds = constpkg.MetricProviderRequestDurationSeconds
)

const (
	opListZones      = "list_zones"
	opListDNSRecords = "list_dns_records"
	opCreateDNS      = "create_dns_record"
	opUpdateDNS      = "update_dns_record"
	opDeleteDNS      = "delete_dns_record"
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

	cf := &Cloudflare{
		AbstractManagedType: adapter.NewManagedType(ProviderType, option.Name),
		client:              internal.NewClient(httpx.NewClient(clientOptions...), option.Token),
		zones:               new(generic.SyncMap[string, string]),

		proxied: option.Proxy,
		logger:  logger,
	}

	return cf, nil
}

type Cloudflare struct {
	adapter.AbstractManagedType

	logger *zap.Logger
	client *internal.Client

	requestTotal    metrics.CounterVec
	requestFailures metrics.CounterVec
	requestDuration metrics.HistogramVec

	zones     *generic.SyncMap[string, string]
	zoneMutex sync.Mutex

	proxied      bool
	privateRoute bool
}

func (c *Cloudflare) Start(ctx context.Context, stage services.Stage) error {
	if stage != services.StagePreStart {
		return nil
	}
	factory := ddnsmetric.FromContext(ctx, c)
	if factory == nil {
		return errors.New("ddnsmetric: provider factory not in context")
	}
	labels := []string{constpkg.MetricLabelName, constpkg.MetricLabelOperation}
	c.requestTotal = factory.CounterVec(metricRequestTotal,
		"Total provider API requests.", labels)
	c.requestFailures = factory.CounterVec(metricRequestFailureTotal,
		"Failed provider API requests.", labels)
	c.requestDuration = factory.HistogramVec(metricRequestDurationSeconds,
		"Provider API request duration.", labels, nil)
	return nil
}

func (c *Cloudflare) Close() error { return nil }

func (c *Cloudflare) recordAPICall(op string, start time.Time, err error) {
	c.requestTotal.With(c.Name(), op).Inc()
	c.requestDuration.With(c.Name(), op).Observe(time.Since(start).Seconds())
	if err != nil {
		c.requestFailures.With(c.Name(), op).Inc()
	}
}

func (c *Cloudflare) getZoneID(ctx context.Context, domain string) (string, error) {
	if !domains.IsDomainName(domain) {
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

func (c *Cloudflare) updateZoneID(ctx context.Context, domain string) (zoneID string, err error) {
	start := time.Now()
	defer func() { c.recordAPICall(opListZones, start, err) }()
	logger := c.logger
	zoneName := c.client.ListZones()

	logger.Info("search zone id from upstream", zap.String("domain", domain))

	for page, perr := zoneName.Next(ctx); perr != io.EOF; page, perr = zoneName.Next(ctx) {
		if perr != nil {
			return "", perr
		}
		for i := 0; i < len(page); i++ {
			zone := page[i]
			logger := logger.WithLazy(
				zap.String("domain", domain),
				zap.String("zone_name", zone.Name),
			)

			if !domains.IsDomainName(zone.Name) {
				logger.Warn("upstream return a bad domain")
				continue
			}

			logger.Info("found zone id")
			if zone.Name == domain {
				zoneID = zone.ID
				c.zones.Store(domain, zone.ID)
			}

			c.zones.Store(zone.Name, zone.ID)
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
	//domainToken := strings.Split(domain, ".")
	//for i := len(domainToken) - 2; i > -1; i-- {
	//	fullDomain := strings.Join(domainToken[i:], ".")
	//	if existedZoneID, existed := c.zones.Load(fullDomain); existed {
	//		return existedZoneID
	//	}
	//}
	cutDomains := domains.CutFromHead(domain)
	for i := 0; i < len(cutDomains); i++ {
		cutDomain := cutDomains[i]
		if existedZoneID, existed := c.zones.Load(cutDomain); existed {
			return existedZoneID
		}
	}
	return ""
}
