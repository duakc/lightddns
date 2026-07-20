package internal

import (
	"context"
	"net"
	"net/netip"
)

func LocalLookup(ctx context.Context, domain string, ipv4, ipv6 bool) ([]netip.Addr, error) {
	switch {
	case ipv4 && ipv6:
		return net.DefaultResolver.LookupNetIP(ctx, "ip", domain)
	case ipv4:
		return net.DefaultResolver.LookupNetIP(ctx, "ip4", domain)
	case ipv6:
		return net.DefaultResolver.LookupNetIP(ctx, "ip6", domain)
	default:
		return nil, nil
	}
}
