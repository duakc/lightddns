package tencentcloud

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/ddnsmetric"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"
	"github.com/duakc/lightddns/infra/netool/resolvectl"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/providers/tencentcloud/internal"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

const ProviderType = constpkg.ProviderTypeTencentCloud

func init() {
	adapter.Register(
		adapter.ProviderRegister,
		ProviderType,
		New,
	)
}

var _ adapter.Provider = (*TencentCloud)(nil)

type TencentCloud struct {
	adapter.AbstractManagedType

	logger *zap.Logger
	client *internal.Client

	// zones memoises the parent zone Name for any FQDN we've already
	// resolved. Tencent's DNSPod APIs key records by Name (no numeric id),
	// so cache value == cache key for the parent zone.
	zones ddnsx.DomainIdCache
}

func New(ctx context.Context, logger *zap.Logger, option options.TencentCloudProviderOption) (adapter.Provider, error) {
	if option.SecretId == "" {
		return nil, fmt.Errorf("tencentcloud(%s): secretId is empty", option.Name)
	}
	if option.SecretKey == "" {
		return nil, fmt.Errorf("tencentcloud(%s): secretKey is empty", option.Name)
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

	return &TencentCloud{
		AbstractManagedType: adapter.NewManagedType(ProviderType, option.Name),
		client: internal.NewClient(logger, option.Name,
			httpx.NewClient(clientOptions...), option.SecretId, option.SecretKey),
		logger: logger,
	}, nil
}

func (t *TencentCloud) RegisterMetrics(factory ddnsmetric.Factory) {
	t.client.RegisterMetrics(factory)
}

func (t *TencentCloud) Close() error { return nil }

// resolveZone returns the parent zone Name for the given FQDN.
func (t *TencentCloud) resolveZone(ctx context.Context, fqdn string) (string, error) {
	if !domains.IsDomainName(fqdn) {
		return "", fmt.Errorf("bad domain name: %s", fqdn)
	}
	zone := t.zones.LoadOrStore(ctx, fqdn, t.client)
	if zone == "" {
		return "", fmt.Errorf("domain not found: %s", fqdn)
	}
	return zone, nil
}

// subDomain returns the host part of fqdn relative to zone. For an apex
// record (fqdn == zone) it returns "@", matching Tencent's API convention.
func subDomain(fqdn, zone string) string {
	if strings.EqualFold(fqdn, zone) {
		return "@"
	}
	if strings.HasSuffix(strings.ToLower(fqdn), "."+strings.ToLower(zone)) {
		return fqdn[:len(fqdn)-len(zone)-1]
	}
	return fqdn
}

func (t *TencentCloud) Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error) {
	diffs, err := t.diff(ctx, domain, addr)
	if err != nil {
		return false, err
	}
	return len(diffs) > 0, nil
}

func (t *TencentCloud) diff(ctx context.Context, domain string, addr []netip.Addr) ([]ddnsx.Diff[internal.Record], error) {
	zone, err := t.resolveZone(ctx, domain)
	if err != nil {
		return nil, err
	}
	sub := subDomain(domain, zone)

	return ddnsx.BuildDiffs(ctx, domain, addr, func(ctx context.Context, _, dnsType string) ([]ddnsx.Existing[internal.Record], error) {
		return t.listExisting(ctx, zone, sub, dnsType)
	})
}

func (t *TencentCloud) listExisting(ctx context.Context, zone, sub, dnsType string) ([]ddnsx.Existing[internal.Record], error) {
	records, err := t.client.DescribeRecordList(ctx, zone, sub, dnsType)
	if err != nil {
		return nil, fmt.Errorf("DescribeRecordList: %w", err)
	}
	var existing []ddnsx.Existing[internal.Record]
	for _, r := range records {
		// Records of different types share the same SubDomain space; the API
		// filter by RecordType should already constrain this, but guard
		// defensively against any leakage.
		if r.Type != dnsType {
			continue
		}
		ip, perr := netip.ParseAddr(r.Value)
		if perr != nil {
			return nil, fmt.Errorf("record %d: not an address: %s: %w", r.RecordId, r.Value, perr)
		}
		existing = append(existing, ddnsx.Existing[internal.Record]{
			Addr:   ip,
			Record: r,
		})
	}
	return existing, nil
}

func (t *TencentCloud) Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (bool, error) {
	logger := t.logger.With(zap.String("domain", domain))
	logger.Debug("new update request", zap.Stringers("addresses", addr))

	zone, err := t.resolveZone(ctx, domain)
	if err != nil {
		return false, err
	}
	sub := subDomain(domain, zone)

	diffs, err := t.diff(ctx, domain, addr)
	if err != nil {
		return false, fmt.Errorf("diff: %w", err)
	}
	if len(diffs) == 0 {
		logger.Info("no difference since last updated, skip")
		return false, nil
	}

	for _, d := range diffs {
		if err := t.applyDiff(ctx, logger, zone, sub, ttl, d); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (t *TencentCloud) applyDiff(ctx context.Context, logger *zap.Logger,
	zone, sub string, ttl uint32, d ddnsx.Diff[internal.Record],
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
		return t.client.CreateRecord(ctx, internal.CreateRecordRequest{
			Domain:     zone,
			SubDomain:  sub,
			RecordType: recordTypeOf(d.Target),
			RecordLine: internal.DefaultRecordLine,
			Value:      d.Target.Unmap().String(),
			TTL:        ttl,
		})
	case ddnsx.DDNSActionUpdate:
		logger.Info("update")
		return t.client.ModifyRecord(ctx, internal.ModifyRecordRequest{
			Domain:     zone,
			RecordId:   d.Record.RecordId,
			SubDomain:  sub,
			RecordType: recordTypeOf(d.Target),
			RecordLine: lineOrDefault(d.Record.Line),
			Value:      d.Target.Unmap().String(),
			TTL:        ttl,
		})
	case ddnsx.DDNSActionDelete:
		logger.Info("delete")
		return t.client.DeleteRecord(ctx, zone, d.Record.RecordId)
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
