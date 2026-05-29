package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/ddnsmetric"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/httpxx"
	"github.com/duakc/lightddns/infra/metrics"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/resolvectl"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/providers/cloudflare/internal"

	"github.com/duakc/mt"
	"github.com/duakc/mt/common/generic"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

// Metric leaf names. Final names get the "ddns_provider_" prefix from the
// ProviderFactory in PreStart.
const (
	metricRequestTotal           = "request_total"
	metricRequestFailureTotal    = "request_failure_total"
	metricRequestDurationSeconds = "request_duration_seconds"
)

// operation label values for the request counters / histogram.
const (
	opListZones      = "list_zones"
	opListDNSRecords = "list_dns_records"
	opCreateDNS      = "create_dns_record"
	opUpdateDNS      = "update_dns_record"
	opDeleteDNS      = "delete_dns_record"
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
		return nil, fmt.Errorf("cloudflare(%s): %w", option.Name, adapter.ErrRequireToken)
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
		httpxx.ClientOptionWithToken(option.Token),
		httpxx.ClientOptionWithDialer(
			resolvectl.NewDialer(ctx, connectDialer,
				mt.Must(option.DNS.NewTransport(ctx, connectDialer)), resolvectl.DefaultResolveClient)))

	cf := &Cloudflare{
		client: internal.NewClient(ctx, httpxx.NewClient(clientOptions...)),
		zones:  new(generic.SyncMap[string, string]),

		name:    option.Name,
		proxied: option.Proxy,
	}

	cf.logger = adapter.CreatProviderLogger(zaplog.FromContext(ctx), cf)

	return cf, nil
}

type Cloudflare struct {
	logger *zap.Logger
	client *internal.Client

	requestTotal    metrics.CounterVec
	requestFailures metrics.CounterVec
	requestDuration metrics.HistogramVec

	name      string
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

// recordAPICall is the per-impl bookkeeping helper: takes captured start time
// and the call's error by value (no pointers).
func (c *Cloudflare) recordAPICall(op string, start time.Time, err error) {
	c.requestTotal.With(c.name, op).Inc()
	c.requestDuration.With(c.name, op).Observe(time.Since(start).Seconds())
	if err != nil {
		c.requestFailures.With(c.name, op).Inc()
	}
}

func (c *Cloudflare) Type() string {
	return constpkg.ProviderTypeCloudflare
}

func (c *Cloudflare) Name() string {
	return c.name
}

func (c *Cloudflare) getZoneID(ctx context.Context, domain string) (string, error) {
	if !netool.IsDomainName(domain) {
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
			if !netool.IsDomainName(zone.Name) {
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
	domainToken := strings.Split(domain, ".")
	for i := len(domainToken) - 2; i > -1; i-- {
		fullDomain := strings.Join(domainToken[i:], ".")
		if existedZoneID, existed := c.zones.Load(fullDomain); existed {
			return existedZoneID
		}
	}
	return ""
}
