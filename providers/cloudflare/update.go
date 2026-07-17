package cloudflare

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter/ddnsx"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/providers/cloudflare/internal"

	"go.uber.org/zap"
)

func (c *Cloudflare) Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (bool, error) {
	logger := c.logger.With(
		zap.String("domain", domain),
	)

	logger.Debug("new update request",
		zap.Stringers("addresses", addr),
	)

	zoneID, err := c.getZoneID(ctx, domain)
	if err != nil {
		return false, fmt.Errorf("getZoneID: %w", err)
	}
	diffs, err := c.diff(ctx, domain, addr)
	if err != nil {
		return false, fmt.Errorf("diff: %w", err)
	}
	if len(diffs) == 0 {
		logger.Info("no difference since last updated, skip")
		return false, nil
	}

	for _, d := range diffs {
		if err := c.applyDiff(ctx, logger, zoneID, ttl, d); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (c *Cloudflare) applyDiff(ctx context.Context, logger *zap.Logger, zoneID string, ttl uint32, d ddnsx.Diff[internal.DNSRecord]) error {
	logFields := []zap.Field{
		zap.String("domain", d.Domain),
		zap.Stringer("source", d.Source),
		zap.Stringer("target", d.Target),
		zap.Stringer("action", d.Action),
	}

	logger = logger.WithLazy(logFields...)
	switch d.Action {
	case ddnsx.DDNSActionCreate:
		logger.Info("create")
		return c.client.CreateDNSRecords(ctx, zoneID, ipToUpdateDNSRecord(d.Domain, d.Target, ttl, c.privateRoute, c.proxied))
	case ddnsx.DDNSActionUpdate:
		logger.Info("update")
		return c.client.UpdateDNSRecords(ctx, zoneID, d.Record.ID, ipToUpdateDNSRecord(d.Domain, d.Target, ttl, c.privateRoute, c.proxied))
	case ddnsx.DDNSActionDelete:
		logger.Info("delete")
		return c.client.DeleteDNSRecord(ctx, zoneID, d.Record.ID)
	}
	return nil
}

func ipToUpdateDNSRecord(name string, ip netip.Addr, ttl uint32, PrivateRouting bool, Proxied bool) internal.UpdateDNSRecordRequest {
	req := internal.UpdateDNSRecordRequest{
		Comment: constpkg.ThisRecordIsManagedByLightddns,

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
