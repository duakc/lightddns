package filter

import (
	"net/netip"

	"go4.org/netipx"
)

type Rule struct {
	Prefixes *netipx.IPSet

	Invert bool
}

func (r Rule) Match(ip netip.Addr) bool {
	if r.Invert {
		return !r.matchInternal(ip)
	}
	return r.matchInternal(ip)
}

func (r Rule) matchInternal(ip netip.Addr) bool {
	if r.Prefixes != nil && !r.Prefixes.Contains(ip) {
		return false
	}

	return true
}
