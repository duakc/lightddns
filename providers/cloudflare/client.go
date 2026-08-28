package cloudflare

import (
	"cmp"
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/adapter/providerx"
	"github.com/duakc/lightddns/infra/netx/domains"
	"github.com/duakc/lightddns/infra/zaplog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	pageSize = 50
)

type ComparedRecord struct {
	Record         Record
	Addr           netip.Addr
	TTL            uint32
	Proxied        bool
	PrivateRouting bool
}

func (r ComparedRecord) Address() netip.Addr { return r.Addr }

func (r ComparedRecord) Compare(other ComparedRecord) int {
	if c := r.Addr.Compare(other.Addr); c != 0 {
		return c
	}
	if c := cmp.Compare(r.TTL, other.TTL); c != 0 {
		return c
	}
	if r.Proxied != other.Proxied {
		if r.Proxied {
			return 1
		}
		return -1
	}
	if r.PrivateRouting == other.PrivateRouting {
		return 0
	}
	if r.PrivateRouting {
		return 1
	}
	return -1
}

func (r ComparedRecord) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", r.Addr.String())
	enc.AddUint32("ttl", r.TTL)
	enc.AddBool("proxied", r.Proxied)
	enc.AddBool("private_routing", r.PrivateRouting)
	if r.Record.ID != "" {
		enc.AddString("record_id", r.Record.ID)
	}
	return nil
}

var (
	_ ddnsx.DDNSClient[ComparedRecord] = (*Client)(nil)
	_ ddnsx.ZoneSearcher               = (*Client)(nil)
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

func (c *Client) Records(ctx context.Context, key ddnsx.RecordKey) ([]ComparedRecord, error) {
	name := domains.FqdnToDomain(key.FQDN)

	var existing []ComparedRecord
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

		Ttl:     desired.TTL,
		Type:    target.Type.String(),
		Content: desired.Addr.Unmap().String(),

		PrivateRouting: desired.PrivateRouting,
		Proxied:        desired.Proxied,
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
