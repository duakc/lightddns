package netool

import "github.com/duakc/lightddns/infra/netool/internal"

var (
	IsIPv4           = internal.IsIPv4
	IsIPv6           = internal.IsIPv6
	FilterAddress    = internal.FilterAddress
	SplitIPv4AndIPv6 = internal.SplitIPv4AndIPv6
)

const (
	DefaultHappyEyeballFallbackDelay = internal.DefaultHappyEyeballFallbackDelay
	DefaultDNSTTL                    = internal.DefaultDNSTTL
)
