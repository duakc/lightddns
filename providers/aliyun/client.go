package aliyun

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"strings"

	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/infra/netx/domains"
	"github.com/duakc/lightddns/infra/numcompare"
	"github.com/duakc/lightddns/infra/zaplog"

	mDns "github.com/miekg/dns"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ComparedRecord struct {
	Record Record
	Addr   netip.Addr
	TTL    uint32
	Line   string
}

var ErrDefaultRecordLineRequired = errors.New("aliyun default record line is required")

func (r ComparedRecord) Address() netip.Addr { return r.Addr }

func (r ComparedRecord) Compare(other ComparedRecord) int {
	if c := cmp.Compare(lineRank(r.Line), lineRank(other.Line)); c != 0 {
		return c
	}
	if c := r.Addr.Compare(other.Addr); c != 0 {
		return c
	}
	return numcompare.TTL(r.TTL, other.TTL)
}

func lineRank(line string) string {
	if line == DefaultRecordLine {
		return ""
	}
	return line
}

func (r ComparedRecord) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", r.Addr.String())
	enc.AddUint32("ttl", r.TTL)
	enc.AddString("line", r.Line)
	if r.Record.RecordId != "" {
		enc.AddString("record_id", r.Record.RecordId)
	}
	return nil
}

type ComparedRecordDDNSClient ddnsx.DDNSClient[ComparedRecord]

// Client implements ddnsx.DDNSClient[ComparedRecord] and ddnsx.ZoneSearcher.
var (
	_ ComparedRecordDDNSClient = (*Client)(nil)
	_ ddnsx.ZoneSearcher       = (*Client)(nil)
)

type Client struct {
	logger *zap.Logger
	api    APIClient
	zones  ddnsx.ZoneCache
	lines  []string
}

func NewClient(logger *zap.Logger, api APIClient, lines ...[]string) *Client {
	var configured []string
	if len(lines) > 0 {
		configured = configuredLines(lines[0])
	} else {
		configured = configuredLines(nil)
	}
	return &Client{
		logger: zaplog.DoNotPanic(logger).Named("client"),
		api:    api,
		lines:  configured,
	}
}

func (c *Client) BuildDiffs(ctx context.Context, key ddnsx.RecordKey, target []netip.Addr,
	ttl uint32, reader ddnsx.RecordReader[ComparedRecord],
) ([]ddnsx.Diff[ComparedRecord], error) {
	lines := c.lines
	return ddnsx.BuildDiffsWith(ctx, key, target, ttl, reader,
		func(key ddnsx.RecordKey, existing []ComparedRecord, target []netip.Addr, ttl uint32) ([]ddnsx.Diff[ComparedRecord], error) {
			if len(existing) == 0 && !slices.Contains(lines, DefaultRecordLine) {
				return nil, fmt.Errorf("%w: the %q line record is required during initialization; add %q to the provider's \"lines\" configuration and create that default record first", ErrDefaultRecordLineRequired, DefaultRecordLine, DefaultRecordLine)
			}
			var diffs []ddnsx.Diff[ComparedRecord]
			for _, line := range lines {
				var current []ComparedRecord
				for _, record := range existing {
					if record.Line == line {
						current = append(current, record)
					}
				}
				want := make([]ComparedRecord, len(target))
				for i, addr := range target {
					want[i] = ComparedRecord{Addr: addr, TTL: ttl, Line: line}
				}
				diffs = append(diffs, ddnsx.Compare(key.FQDN, current, want)...)
			}
			return diffs, nil
		})
}

func (c *Client) ResolveZone(ctx context.Context, fqdn string) (ddnsx.Zone, error) {
	return c.zones.Resolve(ctx, fqdn, c)
}

func (c *Client) SearchZones(ctx context.Context, keyword string) ([]ddnsx.Zone, error) {
	const pageSize = 100

	logger := c.logger.With(
		zap.String("action", AlidnsActionDescribeDomains),
		zap.String("keyword", keyword),
	)
	logger.Info("search domain from upstream")

	var zones []ddnsx.Zone
	for pageNumber := 1; ; pageNumber++ {
		page, err := c.api.DescribeDomains(ctx, DescribeDomainsRequest{
			KeyWord:    domains.FqdnToDomain(keyword),
			PageNumber: pageNumber,
			PageSize:   pageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("DescribeDomains: %w", err)
		}
		for _, domain := range page.Domains.Domain {
			if !domains.IsDomainName(domain.DomainName) {
				continue
			}
			zones = append(zones, ddnsx.Zone{
				Fqdn: domain.DomainName,
				ID:   domain.DomainId,
			})
		}
		if len(page.Domains.Domain) < pageSize ||
			int64(pageNumber)*pageSize >= page.TotalCount {
			return zones, nil
		}
	}
}

func (c *Client) Records(ctx context.Context, key ddnsx.RecordKey) ([]ComparedRecord, error) {
	const pageSize = 100

	rr, err := relativeRR(key.FQDN, key.Zone.Fqdn)
	if err != nil {
		return nil, err
	}

	var records []ComparedRecord
	for pageNumber := 1; ; pageNumber++ {
		page, err := c.api.DescribeDomainRecords(ctx, DescribeDomainRecordsRequest{
			DomainName:  domains.FqdnToDomain(key.Zone.Fqdn),
			RRKeyWord:   rr,
			TypeKeyWord: key.Type.String(),
			PageNumber:  pageNumber,
			PageSize:    pageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("DescribeDomainRecords: %w", err)
		}
		for _, record := range page.DomainRecords.Record {
			if !strings.EqualFold(record.RR, rr) || record.Type != key.Type.String() {
				continue
			}
			address, err := netip.ParseAddr(record.Value)
			if err != nil {
				return nil, fmt.Errorf("DescribeDomainRecords: record %s: not an address: %s: %w", record.RecordId, record.Value, err)
			}
			records = append(records, ComparedRecord{
				Record: record,
				Addr:   address.Unmap(),
				TTL:    record.TTL,
				Line:   cmp.Or(record.Line, DefaultRecordLine),
			})
		}
		if len(page.DomainRecords.Record) < pageSize ||
			int64(pageNumber)*pageSize >= page.TotalCount {
			if len(records) > 0 && !slices.ContainsFunc(records, func(record ComparedRecord) bool {
				return record.Line == DefaultRecordLine
			}) {
				return nil, fmt.Errorf("%w: upstream returned records without the %q line; create that default record first", ErrDefaultRecordLineRequired, DefaultRecordLine)
			}
			return records, nil
		}
	}
}

func (c *Client) Create(ctx context.Context, target ddnsx.RecordSpec, desired ComparedRecord) error {
	rr, err := relativeRR(target.FQDN, target.Zone.Fqdn)
	if err != nil {
		return err
	}
	_, err = c.api.AddDomainRecord(ctx, AddDomainRecordRequest{
		DomainName: domains.FqdnToDomain(target.Zone.Fqdn),
		RR:         rr,
		Type:       target.Type.String(),
		Value:      desired.Addr.String(),
		TTL:        desired.TTL,

		Line: desired.Line,
	})
	return err
}

func (c *Client) Update(ctx context.Context, target ddnsx.RecordSpec, desired, existing ComparedRecord) error {
	rr, err := relativeRR(target.FQDN, target.Zone.Fqdn)
	if err != nil {
		return err
	}
	_, err = c.api.UpdateDomainRecord(ctx, UpdateDomainRecordRequest{
		RecordId: existing.Record.RecordId,

		Type:  target.Type.String(),
		Value: desired.Addr.String(),
		TTL:   desired.TTL,
		RR:    rr,

		Line: desired.Line,
	})
	return err
}

func (c *Client) Delete(ctx context.Context, _ ddnsx.RecordKey, existing ComparedRecord) error {
	_, err := c.api.DeleteDomainRecord(ctx, DeleteDomainRecordRequest{RecordId: existing.Record.RecordId})
	return err
}

func configuredLines(lines []string) []string {
	if len(lines) == 0 {
		return []string{DefaultRecordLine}
	}

	result := make([]string, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		result = append(result, line)
	}

	slices.SortStableFunc(result, func(a, b string) int {
		if a == DefaultRecordLine {
			return -1
		}
		if b == DefaultRecordLine {
			return 1
		}
		return cmp.Compare(a, b)
	})
	if len(result) == 0 {
		return []string{DefaultRecordLine}
	}
	return result
}

func relativeRR(fqdn, zone string) (string, error) {
	fqdn = mDns.Fqdn(fqdn)
	zone = mDns.Fqdn(zone)
	if fqdn == zone {
		return ApexRecordHost, nil
	}
	if rr, found := strings.CutSuffix(fqdn, "."+zone); found {
		return rr, nil
	}
	return "", fmt.Errorf("fqdn %q is not within zone %q", fqdn, zone)
}
