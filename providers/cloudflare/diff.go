package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/netip"
	"time"

	"github.com/duakc/lightddns/infra/ddnsx"
	"github.com/duakc/lightddns/providers/cloudflare/internal"
)

func (c *Cloudflare) Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error) {
	diffs, err := c.diff(ctx, domain, addr)
	if err != nil {
		return false, err
	}
	return len(diffs) > 0, nil
}

func (c *Cloudflare) diff(ctx context.Context, domain string, addr []netip.Addr) ([]ddnsx.Diff[internal.DNSRecord], error) {
	zoneID, err := c.getZoneID(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("getZoneID: %w", err)
	}
	return ddnsx.Build(ctx, domain, addr, func(ctx context.Context, _, dnsType string) ([]ddnsx.Existing[internal.DNSRecord], error) {
		return c.listExisting(ctx, domain, zoneID, dnsType)
	})
}

func (c *Cloudflare) listExisting(ctx context.Context, domain, zoneID, dnsType string) (existing []ddnsx.Existing[internal.DNSRecord], err error) {
	start := time.Now()
	defer func() { c.recordAPICall(opListDNSRecords, start, err) }()

	pager, err := c.client.ListDNSRecords(domain, zoneID, dnsType)
	if err != nil {
		return nil, fmt.Errorf("ListDNSRecords: %w", err)
	}
	for page, perr := pager.Next(ctx); perr != io.EOF; page, perr = pager.Next(ctx) {
		if perr != nil {
			return nil, perr
		}
		for i := 0; i < len(page); i++ {
			record := page[i]
			ip, err := netip.ParseAddr(record.Content)
			if err != nil {
				return nil, fmt.Errorf("not an address: %s: %w", record.Content, err)
			}
			existing = append(existing, ddnsx.Existing[internal.DNSRecord]{
				Addr:   ip,
				Record: record,
			})
		}
	}
	return existing, nil
}
