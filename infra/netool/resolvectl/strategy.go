package resolvectl

import "fmt"

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
