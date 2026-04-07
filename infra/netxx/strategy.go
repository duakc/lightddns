package netxx

import "fmt"

type DialStrategy uint

const (
	DialPreferIPv6 DialStrategy = iota
	DialPreferIPv4
	DialOnlyIPv6
	DialOnlyIPv4
)

func (ds DialStrategy) String() string {
	switch ds {
	case DialPreferIPv6:
		return "prefer_ipv6"
	case DialPreferIPv4:
		return "prefer_ipv4"
	case DialOnlyIPv6:
		return "ipv6_only"
	case DialOnlyIPv4:
		return "ipv4_only"
	default:
		return fmt.Sprintf("DialStrategy_%d", ds)
	}
}

func DialStrategyFromString(s string) (DialStrategy, bool) {
	switch s {
	case "", "prefer_ipv6":
		return DialPreferIPv6, true
	case "prefer_ipv4":
		return DialPreferIPv4, true
	case "ipv6_only":
		return DialOnlyIPv6, true
	case "ipv4_only":
		return DialOnlyIPv4, true
	default:
		return DialPreferIPv6, false
	}
}

type ResolveStrategy uint

const (
	ResolveAsis = iota
	ResolveIPv6
	ResolveIPv4
)

func (rs ResolveStrategy) String() string {
	switch rs {
	case ResolveAsis:
		return "asis"
	case ResolveIPv6:
		return "ipv6_only"
	case ResolveIPv4:
		return "ipv4_only"
	default:
		return fmt.Sprintf("ResolveStrategy_%d", rs)
	}
}
