package netxx

import "net/netip"

func IsIPv4(ip netip.Addr) bool {
	return ip.Is4() || ip.Is4In6()
}

func IsIPv6(ip netip.Addr) bool {
	return ip.Is6() && !ip.Is4In6()
}
