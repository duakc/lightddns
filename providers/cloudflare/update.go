package cloudflare

import (
	"context"
	"fmt"
	"net/netip"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/netxx"
	"github.com/duakc/lightddns/providers/cloudflare/internal"
)

func (c *Cloudflare) Update(ctx context.Context, domain string, ttl int, addr []netip.Addr) error {
	if ttl < 0 {
		ttl = 0
	}
	zoneID, err := c.getZoneID(ctx, domain)
	if err != nil {
		return fmt.Errorf("getZoneID: %w", err)
	}
	diffRecords, err := c.diff(ctx, domain, addr)
	if diff, er := isDiff(diffRecords, err); !diff || er != nil {
		return fmt.Errorf("diff: %w", err)
	}
	for i := 0; i < len(diffRecords); i++ {
		var err error
		rc := diffRecords[i]
		updateRequest := ipToUpdateDNSRecord(domain, rc.address, ttl, c.privateRoute, c.proxied)
		switch {
		case rc.toCreate:
			err = c.client.CreateDNSRecords(ctx, zoneID, updateRequest)
		case rc.toUpdate:
			err = c.client.UpdateDNSRecords(ctx, zoneID, rc.ID, updateRequest)
		case rc.toDelete:
			err = c.client.DeleteDNSRecord(ctx, zoneID, rc.ID)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func ipToUpdateDNSRecord(name string, ip netip.Addr, ttl int, PrivateRouting bool, Proxied bool) internal.UpdateDNSRecordRequest {
	req := internal.UpdateDNSRecordRequest{
		Comment:        "This Record is Managed By Lightddns",
		Name:           name,
		Ttl:            ttl,
		Content:        ip.Unmap().String(),
		PrivateRouting: PrivateRouting,
		Proxied:        Proxied,
	}
	if netxx.IsIPv6(ip) {
		req.Type = constpkg.DNSTypeAAAA
	} else {
		req.Type = constpkg.DNSTypeA
	}
	return req
}
