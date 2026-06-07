package aliyun

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"
	"github.com/duakc/lightddns/infra/netool/resolvectl"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/providers/aliyun/internal"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

const ProviderType = constpkg.ProviderTypeAliyun

var _ adapter.Provider = (*Aliyun)(nil)

func init() {
	adapter.Register(
		adapter.ProviderRegister,
		ProviderType,
		New,
	)
}

func New(ctx context.Context, logger *zap.Logger, option options.AliyunProviderOption) (adapter.Provider, error) {
	if option.AccessKeyId == "" {
		return nil, fmt.Errorf("aliyun(%s): accessKeyId is empty", option.Name)
	}
	if option.AccessKeySecret == "" {
		return nil, fmt.Errorf("aliyun(%s): accessKeySecret is empty", option.Name)
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
	resolveDialer := resolvectl.NewDialer(connectDialer,
		mt.Must(option.DNS.NewTransport(ctx, connectDialer)),
		resolvectl.DefaultResolveClient)

	clientOptions = append(clientOptions,
		httpx.ClientOptionWithDialer(resolveDialer))

	return &Aliyun{
		AbstractManagedType: adapter.NewManagedType(ProviderType, option.Name),
		client: internal.NewClient(ctx, logger, option.Name,
			httpx.NewClient(clientOptions...),
			option.AccessKeyId, option.AccessKeySecret, ""),
		logger: logger.Named("client"),
	}, nil
}

type Aliyun struct {
	adapter.AbstractManagedType

	logger *zap.Logger
	client *internal.Client

	// zones memoises the parent zone name for any FQDN we've already
	// resolved. Aliyun's APIs key zones by DomainName (record APIs don't need
	// a numeric id), so cache value == cache key for the parent zone.
	zones ddnsx.DomainIdCache
}

// resolveZone returns the parent zone name for the given FQDN.
func (a *Aliyun) resolveZone(ctx context.Context, fqdn string) (string, error) {
	if !domains.IsDomainName(fqdn) {
		return "", fmt.Errorf("bad domain name: %s", fqdn)
	}
	zone := a.zones.LoadOrStore(ctx, fqdn, a.client)
	if zone == "" {
		return "", fmt.Errorf("domain not found: %s", fqdn)
	}
	return zone, nil
}

func (a *Aliyun) Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error) {
	diffs, err := a.diff(ctx, domain, addr)
	if err != nil {
		return false, err
	}
	return len(diffs) > 0, nil
}

func (a *Aliyun) diff(ctx context.Context, domain string, addr []netip.Addr) ([]ddnsx.Diff[internal.Record], error) {
	zone, err := a.resolveZone(ctx, domain)
	if err != nil {
		return nil, err
	}
	rr := subDomain(domain, zone)

	return ddnsx.BuildDiffs(ctx, domain, addr, func(ctx context.Context, _, dnsType string) ([]ddnsx.Existing[internal.Record], error) {
		return a.listExisting(ctx, zone, rr, dnsType)
	})
}

func (a *Aliyun) listExisting(ctx context.Context, zone, rr, dnsType string) ([]ddnsx.Existing[internal.Record], error) {
	const pageSize = 100
	var existing []ddnsx.Existing[internal.Record]
	for pageNumber := 1; ; pageNumber++ {
		page, err := a.client.DescribeDomainRecords(ctx, internal.DescribeDomainRecordsRequest{
			DomainName:  zone,
			RRKeyWord:   rr,
			TypeKeyWord: dnsType,
			PageNumber:  pageNumber,
			PageSize:    pageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("DescribeDomainRecords: %w", err)
		}
		for _, r := range page.DomainRecords.Record {
			// RRKeyWord / TypeKeyWord are server-side filters, but Aliyun
			// matches RRKeyWord by substring rather than exact equality —
			// guard so a sibling host like "foo-www" doesn't shadow "www".
			if r.Type != dnsType || !strings.EqualFold(r.RR, rr) {
				continue
			}
			ip, perr := netip.ParseAddr(r.Value)
			if perr != nil {
				return nil, fmt.Errorf("record %s: not an address: %s: %w", r.RecordId, r.Value, perr)
			}
			existing = append(existing, ddnsx.Existing[internal.Record]{
				Addr:   ip,
				Record: r,
			})
		}
		if len(page.DomainRecords.Record) < pageSize ||
			int64(pageNumber)*pageSize >= page.TotalCount {
			return existing, nil
		}
	}
}

func (a *Aliyun) Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (bool, error) {
	logger := a.logger.With(zap.String("domain", domain))
	logger.Debug("new update request", zap.Stringers("addresses", addr))

	zone, err := a.resolveZone(ctx, domain)
	if err != nil {
		return false, err
	}
	rr := subDomain(domain, zone)

	diffs, err := a.diff(ctx, domain, addr)
	if err != nil {
		return false, fmt.Errorf("diff: %w", err)
	}
	if len(diffs) == 0 {
		logger.Info("no difference since last updated, skip")
		return false, nil
	}

	for _, d := range diffs {
		if err := a.applyDiff(ctx, logger, zone, rr, ttl, d); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (a *Aliyun) applyDiff(ctx context.Context, logger *zap.Logger,
	zone, rr string, ttl uint32, d ddnsx.Diff[internal.Record],
) error {
	logFields := []zap.Field{
		zap.String("domain", d.Domain),
		zap.Stringer("action", d.Action),
	}
	if d.Target.IsValid() {
		logFields = append(logFields, zap.Stringer("target", d.Target))
	}
	if d.Source.IsValid() {
		logFields = append(logFields, zap.Stringer("source", d.Source))
	}
	logger = logger.WithLazy(logFields...)
	switch d.Action {
	case ddnsx.DDNSActionCreate:
		logger.Info("create")
		_, err := a.client.AddDomainRecord(ctx, internal.AddDomainRecordRequest{
			DomainName: zone,
			RR:         rr,
			Type:       recordTypeOf(d.Target),
			Line:       internal.DefaultRecordLine,
			Value:      d.Target.Unmap().String(),
			TTL:        ttl,
		})
		return err
	case ddnsx.DDNSActionUpdate:
		logger.Info("update")
		_, err := a.client.UpdateDomainRecord(ctx, internal.UpdateDomainRecordRequest{
			RecordId: d.Record.RecordId,
			RR:       rr,
			Type:     recordTypeOf(d.Target),
			Line:     lineOrDefault(d.Record.Line),
			Value:    d.Target.Unmap().String(),
			TTL:      ttl,
		})
		return err
	case ddnsx.DDNSActionDelete:
		logger.Info("delete")
		_, err := a.client.DeleteDomainRecord(ctx, internal.DeleteDomainRecordRequest{
			RecordId: d.Record.RecordId,
		})
		return err
	}
	return nil
}

func recordTypeOf(ip netip.Addr) string {
	if netool.IsIPv6(ip) {
		return constpkg.DNSTypeAAAA
	}
	return constpkg.DNSTypeA
}

func lineOrDefault(line string) string {
	if line == "" {
		return internal.DefaultRecordLine
	}
	return line
}

// subDomain returns the host part of fqdn relative to zone. For an apex
// record (fqdn == zone) it returns "@", matching Aliyun's RR convention.
func subDomain(fqdn, zone string) string {
	if strings.EqualFold(fqdn, zone) {
		return internal.ApexRecordHost
	}
	if strings.HasSuffix(strings.ToLower(fqdn), "."+strings.ToLower(zone)) {
		return fqdn[:len(fqdn)-len(zone)-1]
	}
	return fqdn
}
