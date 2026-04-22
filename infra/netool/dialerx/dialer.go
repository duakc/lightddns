package dialerx

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"time"

	"github.com/duakc/lightddns/infra/netool/internal"
)

var (
	l4NetworkList = []string{"tcp", "udp", "tcp4", "tcp6", "udp4", "udp6"}
)

var (
	ErrAddressIsDomainName = "address is a domain name"
)

type Dialer interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}

type AddrPortDialer interface {
	Dialer
	DialContextAddrPort(ctx context.Context, network string, address netip.AddrPort) (net.Conn, error)
}

type defaultDialer struct {
	dialer net.Dialer
	bind   *bindDialer

	fallbackDelay time.Duration
	dialStrategy  DialStrategy
}

func (d *defaultDialer) DialContextAddrPort(ctx context.Context, network string, address netip.AddrPort) (net.Conn, error) {
	if d.bind != nil {
		return d.bind.DialContextAddrPort(ctx, network, address)
	}
	return d.dialer.DialContext(ctx, network, address.String())
}

func (d *defaultDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	var this Dialer = &d.dialer
	if d.bind != nil {
		this = d.bind
	}

	if !slices.Contains(l4NetworkList, network) {
		return this.DialContext(ctx, network, address)
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	var addresses []netip.Addr
	var hostPort uint64
	if hostPort, err = strconv.ParseUint(port, 10, 16); err != nil {
		return nil, fmt.Errorf("parse port: %s: %w", port, err)
	}

	if internal.IsDomainName(host) {
		addresses, err = internal.LocalLookup(ctx, host,
			network != "udp6" && network != "tcp6" && d.dialStrategy != DialOnlyIPv6,
			network != "udp4" && network != "tcp4" && d.dialStrategy != DialOnlyIPv4)
		if err != nil {
			return nil, fmt.Errorf("parse address: %s: lookup %s: %w", address, host, err)
		}
	} else {
		var addr netip.Addr
		if addr, err = netip.ParseAddr(host); err != nil {
			return nil, fmt.Errorf("parse address %s: %w", address, err)
		}
		if internal.IsIPv4(addr) && d.dialStrategy != DialOnlyIPv6 ||
			internal.IsIPv6(addr) && d.dialStrategy != DialOnlyIPv4 {
			return nil, fmt.Errorf("dialStrategy=%s,addr=%s: no address to dial",
				d.dialStrategy, addr)
		}
		addresses = []netip.Addr{addr}
	}

	if len(addresses) > 1 && d.fallbackDelay >= 0 {
		return DialParallel(ctx, this, network, addresses, uint16(hostPort),
			d.dialStrategy == DialPreferIPv6, d.fallbackDelay)
	}

	return DialSerial(ctx, this, network, addresses, uint16(hostPort))
}

func NewDialerWithOption(opt ...DialerOption) Dialer {
	d := &defaultDialer{
		dialer:        net.Dialer{},
		fallbackDelay: defaultFallbackDelay,
		dialStrategy:  DialPreferIPv6,
	}

	for i := 0; i < len(opt); i++ {
		o := opt[i]
		o.Apply(d)
	}

	return d
}

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
