package tencentcloud

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/ddnsmetric"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/infra/metrics"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/netool/dialerx"
	"github.com/duakc/lightddns/infra/netool/domains"
	"github.com/duakc/lightddns/infra/netool/httpx"
	"github.com/duakc/lightddns/infra/netool/resolvectl"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"github.com/duakc/lightddns/providers/tencentcloud/internal"

	"github.com/duakc/mt"
	"github.com/duakc/mt/common/generic"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

const (
	metricRequestTotal           = "request_total"
	metricRequestFailureTotal    = "request_failure_total"
	metricRequestDurationSeconds = "request_duration_seconds"
)

const (
	opDescribeDomains = "describe_domains"
	opListRecords     = "list_records"
	opCreateRecord    = "create_record"
	opModifyRecord    = "modify_record"
	opDeleteRecord    = "delete_record"
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
	c      *internal.Client

	requestTotal    metrics.CounterVec
	requestFailures metrics.CounterVec
	requestDuration metrics.HistogramVec

	// domainCache memoises the parent zone (Domain) lookup for an FQDN.
	// Key: any FQDN the caller passed in. Value: the matching DomainInfo.
	domainCache *generic.SyncMap[string, internal.DomainInfo]
	domainMutex sync.Mutex
}

func New(ctx context.Context, option options.TencentCloudProviderOption) (adapter.Provider, error) {
	if option.SecretId == "" || option.SecretKey == "" {
		return nil, fmt.Errorf("tencentcloud(%s): %w", option.Name, adapter.ErrRequireToken)
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
			resolvectl.NewDialer(ctx, connectDialer,
				mt.Must(option.DNS.NewTransport(ctx, connectDialer)), resolvectl.DefaultResolveClient)))

	tc := &TencentCloud{
		AbstractManagedType: adapter.NewManagedType(ProviderType, option.Name),
		c:                   internal.NewClient(httpx.NewClient(clientOptions...), option.SecretId, option.SecretKey),
		domainCache:         new(generic.SyncMap[string, internal.DomainInfo]),
	}
	tc.logger = adapter.CreatProviderLogger(zaplog.FromContext(ctx), tc)
	return tc, nil
}

func (t *TencentCloud) Start(ctx context.Context, stage services.Stage) error {
	if stage != services.StagePreStart {
		return nil
	}
	factory := ddnsmetric.FromContext(ctx, t)
	if factory == nil {
		return errors.New("ddnsmetric: provider factory not in context")
	}
	labels := []string{constpkg.MetricLabelName, constpkg.MetricLabelOperation}
	t.requestTotal = factory.CounterVec(metricRequestTotal,
		"Total provider API requests.", labels)
	t.requestFailures = factory.CounterVec(metricRequestFailureTotal,
		"Failed provider API requests.", labels)
	t.requestDuration = factory.HistogramVec(metricRequestDurationSeconds,
		"Provider API request duration.", labels, nil)
	return nil
}

func (t *TencentCloud) Close() error { return nil }

func (t *TencentCloud) recordAPICall(op string, start time.Time, err error) {
	t.requestTotal.With(t.Name(), op).Inc()
	t.requestDuration.With(t.Name(), op).Observe(time.Since(start).Seconds())
	if err != nil {
		t.requestFailures.With(t.Name(), op).Inc()
	}
}

// resolveDomain returns the parent zone DomainInfo for the given FQDN.
// Results are cached per-FQDN; concurrent first-time lookups are serialised
// behind domainMutex.
func (t *TencentCloud) resolveDomain(ctx context.Context, fqdn string) (internal.DomainInfo, error) {
	if !domains.IsDomainName(fqdn) {
		return internal.DomainInfo{}, fmt.Errorf("bad domain name: %s", fqdn)
	}
	if di, ok := t.domainCache.Load(fqdn); ok {
		return di, nil
	}
	t.domainMutex.Lock()
	defer t.domainMutex.Unlock()
	if di, ok := t.domainCache.Load(fqdn); ok {
		return di, nil
	}

	start := time.Now()
	di, err := t.c.DomainInfo(ctx, fqdn)
	t.recordAPICall(opDescribeDomains, start, err)
	if err != nil {
		return internal.DomainInfo{}, fmt.Errorf("DomainInfo: %w", err)
	}
	t.domainCache.Store(fqdn, di)
	return di, nil
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
	di, err := t.resolveDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	sub := subDomain(domain, di.Name)

	return ddnsx.Build(ctx, domain, addr, func(ctx context.Context, _, dnsType string) ([]ddnsx.Existing[internal.Record], error) {
		return t.listExisting(ctx, di.Name, sub, dnsType)
	})
}

func (t *TencentCloud) listExisting(ctx context.Context, zone, sub, dnsType string) (existing []ddnsx.Existing[internal.Record], err error) {
	start := time.Now()
	defer func() { t.recordAPICall(opListRecords, start, err) }()

	records, err := t.c.DescribeRecordList(ctx, zone, sub, dnsType)
	if err != nil {
		return nil, fmt.Errorf("DescribeRecordList: %w", err)
	}
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

	di, err := t.resolveDomain(ctx, domain)
	if err != nil {
		return false, err
	}
	sub := subDomain(domain, di.Name)

	diffs, err := t.diff(ctx, domain, addr)
	if err != nil {
		return false, fmt.Errorf("diff: %w", err)
	}
	if len(diffs) == 0 {
		logger.Info("no difference since last updated, skip")
		return false, nil
	}

	for _, d := range diffs {
		if err := t.applyDiff(ctx, logger, di.Name, sub, ttl, d); err != nil {
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
		zap.Stringer("source", d.Source),
		zap.Stringer("target", d.Target),
		zap.Stringer("action", d.Action),
	}
	logger = logger.WithLazy(logFields...)
	start := time.Now()
	var err error
	switch d.Action {
	case ddnsx.DDNSActionCreate:
		logger.Info("create")
		err = t.c.CreateRecord(ctx, internal.CreateRecordRequest{
			Domain:     zone,
			SubDomain:  sub,
			RecordType: recordTypeOf(d.Target),
			RecordLine: internal.DefaultRecordLine,
			Value:      d.Target.Unmap().String(),
			TTL:        ttl,
		})
		t.recordAPICall(opCreateRecord, start, err)
	case ddnsx.DDNSActionUpdate:
		logger.Info("update")
		err = t.c.ModifyRecord(ctx, internal.ModifyRecordRequest{
			Domain:     zone,
			RecordId:   d.Record.RecordId,
			SubDomain:  sub,
			RecordType: recordTypeOf(d.Target),
			RecordLine: lineOrDefault(d.Record.Line),
			Value:      d.Target.Unmap().String(),
			TTL:        ttl,
		})
		t.recordAPICall(opModifyRecord, start, err)
	case ddnsx.DDNSActionDelete:
		logger.Info("delete")
		err = t.c.DeleteRecord(ctx, zone, d.Record.RecordId)
		t.recordAPICall(opDeleteRecord, start, err)
	}
	return err
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
