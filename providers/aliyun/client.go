package aliyun

import (
	"cmp"
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/infra/netx/domains"
	"github.com/duakc/lightddns/infra/zaplog"

	mDns "github.com/miekg/dns"
	"go.uber.org/zap"
)

var (
	_ ddnsx.DDNSClient[Record] = (*Client)(nil)
	_ ddnsx.ZoneSearcher       = (*Client)(nil)
)

type Client struct {
	logger *zap.Logger
	api    APIClient
	zones  ddnsx.ZoneCache
}

func NewClient(logger *zap.Logger, api APIClient) *Client {
	return &Client{
		logger: zaplog.DoNotPanic(logger).Named("client"),
		api:    api,
	}
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
			KeyWord:    domains.NormalizeFQDN(keyword),
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

func (c *Client) Records(ctx context.Context, key ddnsx.RecordKey) ([]ddnsx.Existing[Record], error) {
	const pageSize = 100

	rr, err := relativeRR(key.FQDN, key.Zone.Fqdn)
	if err != nil {
		return nil, err
	}

	var records []ddnsx.Existing[Record]
	for pageNumber := 1; ; pageNumber++ {
		page, err := c.api.DescribeDomainRecords(ctx, DescribeDomainRecordsRequest{
			DomainName:  domains.NormalizeFQDN(key.Zone.Fqdn),
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
				return nil, fmt.Errorf("record %s: not an address: %s: %w", record.RecordId, record.Value, err)
			}
			records = append(records, ddnsx.Existing[Record]{
				Addr:   address.Unmap(),
				Record: record,
			})
		}
		if len(page.DomainRecords.Record) < pageSize ||
			int64(pageNumber)*pageSize >= page.TotalCount {
			return records, nil
		}
	}
}

func (c *Client) Create(ctx context.Context, target ddnsx.RecordSpec) error {
	rr, err := relativeRR(target.FQDN, target.Zone.Fqdn)
	if err != nil {
		return err
	}
	_, err = c.api.AddDomainRecord(ctx, AddDomainRecordRequest{
		DomainName: domains.NormalizeFQDN(target.Zone.Fqdn),
		RR:         rr,
		Type:       target.Type.String(),
		Value:      target.Address.String(),
		TTL:        target.TTL,

		Line: DefaultRecordLine,
	})
	return err
}

func (c *Client) Update(ctx context.Context, target ddnsx.RecordSpec, record Record) error {
	rr, err := relativeRR(target.FQDN, target.Zone.Fqdn)
	if err != nil {
		return err
	}
	_, err = c.api.UpdateDomainRecord(ctx, UpdateDomainRecordRequest{
		RecordId: record.RecordId,

		Type:  target.Type.String(),
		Value: target.Address.String(),
		TTL:   target.TTL,
		RR:    rr,

		Line: cmp.Or(record.Line, DefaultRecordLine),
	})
	return err
}

func (c *Client) Delete(ctx context.Context, _ ddnsx.RecordKey, record Record) error {
	_, err := c.api.DeleteDomainRecord(ctx, DeleteDomainRecordRequest{RecordId: record.RecordId})
	return err
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
