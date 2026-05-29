package cloudflare

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/providers/cloudflare/internal"

	"go.uber.org/zap"
)

func (c *Cloudflare) Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (bool, error) {
	logger := c.logger.With(zap.String("domain", domain))
	logger.Debug("new update request",
		zap.Stringers("addresses", addr))

	zoneID, err := c.getZoneID(ctx, domain)
	if err != nil {
		return false, fmt.Errorf("getZoneID: %w", err)
	}
	diffRecords, err := c.diff(ctx, domain, addr)
	if diff, er := isDiff(diffRecords, err); !diff || er != nil {
		if len(diffRecords) == 0 && er == nil {
			logger.Info("no difference since last updated, skip")
			return false, nil
		}
		return false, fmt.Errorf("diff: %w", err)
	}
	for i := 0; i < len(diffRecords); i++ {
		var err error
		rc := diffRecords[i]
		updateRequest := ipToUpdateDNSRecord(domain, rc.address, ttl, c.privateRoute, c.proxied)
		logFields := []zap.Field{
			zap.String("ip", updateRequest.Content),
			zap.String("domain", updateRequest.Name),
		}
		start := time.Now()
		switch {
		case rc.toCreate:
			logger.Info("create", logFields...)
			err = c.client.CreateDNSRecords(ctx, zoneID, updateRequest)
			c.recordAPICall(opCreateDNS, start, err)
		case rc.toUpdate:
			logger.Info("update", logFields...)
			err = c.client.UpdateDNSRecords(ctx, zoneID, rc.ID, updateRequest)
			c.recordAPICall(opUpdateDNS, start, err)
		case rc.toDelete:
			logger.Info("delete", logFields...)
			err = c.client.DeleteDNSRecord(ctx, zoneID, rc.ID)
			c.recordAPICall(opDeleteDNS, start, err)
		}
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func ipToUpdateDNSRecord(name string, ip netip.Addr, ttl uint32, PrivateRouting bool, Proxied bool) internal.UpdateDNSRecordRequest {
	req := internal.UpdateDNSRecordRequest{
		Comment:        "This Record is Managed By Lightddns",
		Name:           name,
		Ttl:            ttl,
		Content:        ip.Unmap().String(),
		PrivateRouting: PrivateRouting,
		Proxied:        Proxied,
	}
	if netool.IsIPv6(ip) {
		req.Type = constpkg.DNSTypeAAAA
	} else {
		req.Type = constpkg.DNSTypeA
	}
	return req
}
