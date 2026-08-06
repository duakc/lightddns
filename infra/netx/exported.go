package netx

import (
	"net/netip"

	"github.com/duakc/lightddns/infra/netx/internal"
)

func IsIPv4(ip netip.Addr) bool {
	return internal.IsIPv4(ip)
}

func IsIPv6(ip netip.Addr) bool {
	return internal.IsIPv6(ip)
}

func FilterAddress(ips []netip.Addr, ipv4, ipv6 bool) []netip.Addr {
	return internal.FilterAddress(ips, ipv4, ipv6)
}

func SplitIPv4AndIPv6(ips []netip.Addr) (ipv4, ipv6 []netip.Addr) {
	return internal.SplitIPv4AndIPv6(ips)
}

func IsBogon(ip netip.Addr) bool {
	return internal.IsBogon(ip)
}

const (
	DefaultHappyEyeballFallbackDelay = internal.DefaultHappyEyeballFallbackDelay
	DefaultDNSTTL                    = internal.DefaultDNSTTL
)
