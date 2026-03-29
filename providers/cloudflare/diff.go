package cloudflare

import (
	"context"
	"fmt"
	"io"
	"net/netip"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/common"
	"github.com/duakc/lightddns/infra/netxx"
	"github.com/duakc/lightddns/providers/cloudflare/internal"
)

type dnsUpdateRequest struct {
	internal.DNSRecord

	address  netip.Addr
	toDelete bool
	toCreate bool
	toUpdate bool
}

func (c *Cloudflare) Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error) {
	return isDiff(c.diff(ctx, domain, addr))
}

func (c *Cloudflare) diff(ctx context.Context, domain string, addr []netip.Addr) ([]dnsUpdateRequest, error) {
	zoneID, err := c.getZoneID(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("getZoneID: %w", err)
	}
	var differentRecords []dnsUpdateRequest
	if ipv4Address := common.Filter(addr, netxx.IsIPv4); len(ipv4Address) > 0 || len(addr) == 0 {
		diffType, err := c.diffType(ctx, domain, zoneID, ipv4Address, constpkg.DNSTypeA)
		if err != nil {
			return nil, err
		}
		differentRecords = append(differentRecords, diffType...)
	}
	if ipv6Address := common.Filter(addr, netxx.IsIPv6); len(ipv6Address) > 0 || len(addr) == 0 {
		diffType, err := c.diffType(ctx, domain, zoneID, ipv6Address, constpkg.DNSTypeAAAA)
		if err != nil {
			return nil, err
		}
		differentRecords = append(differentRecords, diffType...)
	}

	return differentRecords, nil
}

func (c *Cloudflare) diffType(ctx context.Context, domain, zoneID string, addr []netip.Addr, typ string) ([]dnsUpdateRequest, error) {
	records, err := c.client.ListDNSRecords(domain, zoneID, typ)
	if err != nil {
		return nil, fmt.Errorf("ListDNSRecords: %w", err)
	}
	return compareRecords(ctx, addr, records)
}

func compareRecords(ctx context.Context, addresses []netip.Addr, records *internal.PageConfig[internal.DNSRecord]) ([]dnsUpdateRequest, error) {
	var (
		diffRecords  []dnsUpdateRequest
		addressesMap = common.ToMap(addresses)
	)

	for page, err := records.Next(ctx); err != io.EOF; page, err = records.Next(ctx) {
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(page); i++ {
			record := page[i]
			addr, err := netip.ParseAddr(record.Content)
			if err != nil {
				return nil, fmt.Errorf("not a address: %s: %w", record.Content, err)
			}

			if addressesMap[addr] {
				delete(addressesMap, addr)
				continue
			} else if r := newUpdateRequest(record, addr); len(addressesMap) == 0 {
				r.toDelete = true
				diffRecords = append(diffRecords, r)
			} else {
				r.toUpdate = true
				diffRecords = append(diffRecords, r)
			}
		}
	}
	for i := 0; i < len(diffRecords); i++ {
		if diffRecords[i].toUpdate {
			for addr, _ := range addressesMap {
				// pick one and delete
				diffRecords[i].address = addr
				delete(addressesMap, addr)
				break
			}
		}
	}

	for addr, _ := range addressesMap {
		r := newUpdateRequest(internal.DNSRecord{}, addr)
		r.toCreate = true
		diffRecords = append(diffRecords, r)
	}
	return diffRecords, nil
}

func newUpdateRequest(r internal.DNSRecord, addr netip.Addr) dnsUpdateRequest {
	return dnsUpdateRequest{
		DNSRecord: r,
		address:   addr,
	}
}

func isDiff(elements []dnsUpdateRequest, err error) (bool, error) {
	return err == nil && len(elements) > 0, err
}
