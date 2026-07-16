package tencentcloud

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/options/castoption"
	"github.com/duakc/lightddns/providers/tencentcloud/internal"

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
		return nil, fmt.Errorf("secretId is empty")
	}

	if option.SecretKey == "" {
		return nil, fmt.Errorf("secretKey is empty")
	}

	_, _, httpClient, err := castoption.BuildHTTPClientFromScratch(
		logger, option.Connect, option.DNS, option.HTTP)
	if err != nil {
		return nil, err
	}

	return &TencentCloud{
		AbstractManagedType: adapter.NewManagedType(ProviderType, option.Name),
		client: internal.NewClient(ctx, logger, option.Name,
			httpClient, option.SecretId, option.SecretKey),
		logger: logger.Named("client"),
	}, nil
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
	const pageSize = 100
	var existing []ddnsx.Existing[internal.Record]
	for offset := 0; ; {
		page, err := t.client.DescribeRecordList(ctx, internal.DescribeRecordListRequest{
			Domain:     zone,
			Subdomain:  sub,
			RecordType: dnsType,
			Limit:      pageSize,
			Offset:     offset,
		})
		if err != nil {
			// Tencent returns NoDataOfRecord when the domain has no matching
			// records. For initial DDNS setup that's the expected state, not
			// a failure — surface it as an empty list so the caller proceeds
			// to CreateRecord instead of bailing.
			if apiErr, ok := errors.AsType[*internal.APIError](err); ok &&
				apiErr.Code == internal.DNSPodErrCodeNoDataOfRecord {

				return existing, nil
			}
			return nil, fmt.Errorf("DescribeRecordList: %w", err)
		}
		for _, r := range page.RecordList {
			// Records of different types share the same SubDomain space; the
			// API filter by RecordType should already constrain this, but
			// guard defensively against any leakage.
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
		offset += len(page.RecordList)
		if len(page.RecordList) == 0 || offset >= page.RecordCountInfo.TotalCount {
			return existing, nil
		}
	}
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
		_, err := t.client.CreateRecord(ctx, internal.CreateRecordRequest{
			Domain:     zone,
			SubDomain:  sub,
			RecordType: recordTypeOf(d.Target),
			RecordLine: internal.DefaultRecordLine,
			Value:      d.Target.Unmap().String(),
			TTL:        ttl,
		})
		return err
	case ddnsx.DDNSActionUpdate:
		logger.Info("update")
		_, err := t.client.ModifyRecord(ctx, internal.ModifyRecordRequest{
			Domain:     zone,
			RecordId:   d.Record.RecordId,
			SubDomain:  sub,
			RecordType: recordTypeOf(d.Target),
			RecordLine: lineOrDefault(d.Record.Line),
			Value:      d.Target.Unmap().String(),
			TTL:        ttl,
		})
		return err
	case ddnsx.DDNSActionDelete:
		logger.Info("delete")
		_, err := t.client.DeleteRecord(ctx, internal.DeleteRecordRequest{
			Domain:   zone,
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
