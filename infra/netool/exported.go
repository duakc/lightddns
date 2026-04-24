package netool

import "github.com/duakc/lightddns/infra/netool/internal"

var (
	IsSubDomain   = internal.IsSubDomain
	IsDomainName  = internal.IsDomainName
	IsIPv4        = internal.IsIPv4
	IsIPv6        = internal.IsIPv6
	FilterAddress = internal.FilterAddress
)

const (
	DefaultHappyEyeballFallbackDelay = internal.DefaultHappyEyeballFallbackDelay
	DefaultDNSTTL                    = internal.DefaultDNSTTL
)
