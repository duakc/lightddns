package resolvectl

import "fmt"

type LookupStrategy uint

const (
	ResolveAsis LookupStrategy = iota
	ResolveIPv6
	ResolveIPv4
)

func (rs LookupStrategy) String() string {
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
