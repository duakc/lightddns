package cloudflare

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/adapter/providerx"
	"github.com/duakc/lightddns/infra/netx/domains"
	"github.com/duakc/lightddns/infra/zaplog"

	"go.uber.org/zap"
)

const (
	defaultPageSize = 50
)

type ComparedRecordDDNSClient ddnsx.DDNSClient[ComparedRecord]

// Client implements ddnsx.DDNSClient[ComparedRecord] and ddnsx.ZoneSearcher.
var (
	_ ComparedRecordDDNSClient = (*Client)(nil)
	_ ddnsx.ZoneSearcher       = (*Client)(nil)
)

type Client struct {
	logger *zap.Logger
	api    APIClient

	zones ddnsx.ZoneCache

	proxied        bool
	privateRouting bool
}

func NewClient(logger *zap.Logger, api APIClient,
	proxied, privateRouting bool,
) *Client {
	return &Client{
		logger:         zaplog.DoNotPanic(logger).Named("client"),
		api:            api,
		proxied:        proxied,
		privateRouting: privateRouting,
	}
}

func (c *Client) BuildDiffs(ctx context.Context, key ddnsx.RecordKey, target []netip.Addr,
	ttl uint32, reader ddnsx.RecordReader[ComparedRecord],
) ([]ddnsx.Diff[ComparedRecord], error) {
	return ddnsx.BuildDiffs(ctx, key, target, ttl, reader,
		func(addr netip.Addr, ttl uint32) ComparedRecord {
			return ComparedRecord{
				Addr:           addr.Unmap(),
				TTL:            ttl,
				Proxied:        c.proxied,
				PrivateRouting: c.privateRouting,
			}
		})
}

func (c *Client) ResolveZone(ctx context.Context, fqdn string) (ddnsx.Zone, error) {
	return c.zones.Resolve(ctx, fqdn, c)
}

func (c *Client) SearchZones(ctx context.Context, keyword string) ([]ddnsx.Zone, error) {
	var zones []ddnsx.Zone
	for pageNumber := 1; ; pageNumber++ {
		// Cloudflare's name filter won't work for zone discovery, so deliberately
		// ignore keyword and paginate the complete active-zone list.
		page, err := c.api.ListZones(ctx, ListZonesRequest{
			Status:  ZoneStatusActive,
			Page:    pageNumber,
			PerPage: defaultPageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("list zones: %w", err)
		}
		for _, zone := range page.Result {
			if !domains.IsDomainName(zone.Name) {
				c.logger.Warn("upstream returned a bad zone name", zap.String("zone_name", zone.Name))
				continue
			}
			zones = append(zones, ddnsx.Zone{Fqdn: zone.Name, ID: zone.ID})
		}
		if pageDone(page.ResultInfo, pageNumber, len(page.Result)) {
			return zones, nil
		}
	}
}

func (c *Client) Records(ctx context.Context, key ddnsx.RecordKey) ([]ComparedRecord, error) {
	name := domains.FqdnToDomain(key.FQDN)

	var existing []ComparedRecord
	for pageNumber := 1; ; pageNumber++ {
		page, err := c.api.ListDNSRecords(ctx, ListDNSRecordsRequest{
			ZoneID:  key.Zone.ID,
			Name:    name,
			Type:    key.Type.String(),
			Page:    pageNumber,
			PerPage: defaultPageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("list DNS records: %w", err)
		}
		for _, record := range page.Result {
			address, err := netip.ParseAddr(record.Content)
			if err != nil {
				return nil, fmt.Errorf("record %s: not an address: %s: %w", record.ID, record.Content, err)
			}
			existing = append(existing, ComparedRecord{
				Record:         record,
				Addr:           address.Unmap(),
				TTL:            record.Ttl,
				Proxied:        record.Proxied,
				PrivateRouting: record.PrivateRouting,
			})
		}
		if pageDone(page.ResultInfo, pageNumber, len(page.Result)) {
			return existing, nil
		}
	}
}

func (c *Client) Create(ctx context.Context, target ddnsx.RecordSpec, desired ComparedRecord) error {
	body, err := c.recordRequest(target, desired)
	if err != nil {
		return err
	}
	_, err = c.api.CreateDNSRecord(ctx, CreateDNSRecordRequest{
		ZoneID: target.Zone.ID,
		Body:   body,
	})
	return err
}

func (c *Client) Update(ctx context.Context, target ddnsx.RecordSpec, desired, existing ComparedRecord) error {
	body, err := c.recordRequest(target, desired)
	if err != nil {
		return err
	}
	_, err = c.api.UpdateDNSRecord(ctx, UpdateDNSRecordRequest{
		ZoneID:   target.Zone.ID,
		RecordID: existing.Record.ID,
		Body:     body,
	})
	return err
}

func (c *Client) Delete(ctx context.Context, key ddnsx.RecordKey, existing ComparedRecord) error {
	_, err := c.api.DeleteDNSRecord(ctx, DeleteDNSRecordRequest{
		ZoneID:   key.Zone.ID,
		RecordID: existing.Record.ID,
	})
	return err
}

func (c *Client) recordRequest(target ddnsx.RecordSpec, desired ComparedRecord) (DNSRecordRequest, error) {
	name := domains.FqdnToDomain(target.FQDN)

	return DNSRecordRequest{
		Comment: providerx.UpdateMessage(""),

		Name: name,
		Type: target.Type.String(),

		Ttl:            desired.TTL,
		Content:        desired.Addr.Unmap().String(),
		Proxied:        desired.Proxied,
		PrivateRouting: desired.PrivateRouting,
	}, nil
}

func pageDone(info ResultInfo, pageNumber, resultCount int) bool {
	if resultCount == 0 {
		return true
	}
	if info.TotalPages > 0 {
		return pageNumber >= info.TotalPages
	}
	return resultCount < defaultPageSize
}
