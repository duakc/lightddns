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
	pageSize = 50
)

var (
	_ ddnsx.DDNSClient[Record] = (*Client)(nil)
	_ ddnsx.ZoneSearcher       = (*Client)(nil)
)

type Client struct {
	logger *zap.Logger
	api    APIClient

	zones ddnsx.ZoneCache

	proxied        bool
	privateRouting bool
	comment        string
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
			PerPage: pageSize,
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

func (c *Client) Records(ctx context.Context, key ddnsx.RecordKey) ([]ddnsx.Existing[Record], error) {
	name := domains.FqdnToDomain(key.FQDN)

	var existing []ddnsx.Existing[Record]
	for pageNumber := 1; ; pageNumber++ {
		page, err := c.api.ListDNSRecords(ctx, ListDNSRecordsRequest{
			ZoneID:  key.Zone.ID,
			Name:    name,
			Type:    key.Type.String(),
			Page:    pageNumber,
			PerPage: pageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("list DNS records: %w", err)
		}
		for _, record := range page.Result {
			address, err := netip.ParseAddr(record.Content)
			if err != nil {
				return nil, fmt.Errorf("record %s: not an address: %s: %w", record.ID, record.Content, err)
			}
			existing = append(existing, ddnsx.Existing[Record]{
				Addr:   address.Unmap(),
				Record: record,
			})
		}
		if pageDone(page.ResultInfo, pageNumber, len(page.Result)) {
			return existing, nil
		}
	}
}

func (c *Client) Create(ctx context.Context, target ddnsx.RecordSpec) error {
	body, err := c.recordRequest(target)
	if err != nil {
		return err
	}
	_, err = c.api.CreateDNSRecord(ctx, CreateDNSRecordRequest{
		ZoneID: target.Zone.ID,
		Body:   body,
	})
	return err
}

func (c *Client) Update(ctx context.Context, target ddnsx.RecordSpec, record Record) error {
	body, err := c.recordRequest(target)
	if err != nil {
		return err
	}
	_, err = c.api.UpdateDNSRecord(ctx, UpdateDNSRecordRequest{
		ZoneID:   target.Zone.ID,
		RecordID: record.ID,
		Body:     body,
	})
	return err
}

func (c *Client) Delete(ctx context.Context, key ddnsx.RecordKey, record Record) error {
	_, err := c.api.DeleteDNSRecord(ctx, DeleteDNSRecordRequest{
		ZoneID:   key.Zone.ID,
		RecordID: record.ID,
	})
	return err
}

func (c *Client) recordRequest(target ddnsx.RecordSpec) (DNSRecordRequest, error) {
	name := domains.FqdnToDomain(target.FQDN)

	return DNSRecordRequest{
		Comment: providerx.UpdateMessage(""),

		Name: name,

		Ttl:     target.TTL,
		Type:    target.Type.String(),
		Content: target.Address.Unmap().String(),

		PrivateRouting: c.privateRouting,
		Proxied:        c.proxied,
	}, nil
}

func pageDone(info ResultInfo, pageNumber, resultCount int) bool {
	if resultCount == 0 {
		return true
	}
	if info.TotalPages > 0 {
		return pageNumber >= info.TotalPages
	}
	return resultCount < pageSize
}
