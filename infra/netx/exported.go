package netx

import "github.com/duakc/lightddns/infra/netx/internal"

var (
	IsIPv4           = internal.IsIPv4
	IsIPv6           = internal.IsIPv6
	IsBogon          = internal.IsBogon
	FilterAddress    = internal.FilterAddress
	SplitIPv4AndIPv6 = internal.SplitIPv4AndIPv6
)

const (
	DefaultHappyEyeballFallbackDelay = internal.DefaultHappyEyeballFallbackDelay
	DefaultDNSTTL                    = internal.DefaultDNSTTL
)
