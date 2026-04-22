package internal

import (
	"net/netip"

	"github.com/duakc/mt"
)

func IsIPv4(ip netip.Addr) bool {
	return ip.Is4() || ip.Is4In6()
}

func IsIPv6(ip netip.Addr) bool {
	return ip.Is6() && !ip.Is4In6()
}

func FilterAddress(ips []netip.Addr, ipv4, ipv6 bool) []netip.Addr {
	if !ipv4 && !ipv6 {
		if ips == nil {
			return nil
		}
		return []netip.Addr{}
	}

	return mt.Filter(ips, func(v netip.Addr) bool {
		return v.IsValid() && (ipv6 && ipv4 ||
			((ipv4 && IsIPv4(v)) || (ipv6 && IsIPv6(v))))
	})
}
