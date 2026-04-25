package dialerx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"strings"

	"github.com/duakc/lightddns/infra/netool/internal"

	"github.com/duakc/mt"
)

var (
	ErrAddressIsDomainName = errors.New("address is a domain name")
	ErrNoAddressToDialer   = errors.New("no address to dial")
)

type Dialer interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}

type AddrPortDialer interface {
	Dialer
	DialAddrPort(ctx context.Context, network string, addr netip.Addr, port uint16) (net.Conn, error)
}

type ParallelDialer interface {
	Dialer
	DialParallel(ctx context.Context, addresses []netip.Addr, port uint16) (net.Conn, error)
}

type systemDialer struct {
	bindDialer
	happyEyeballConf

	dialer  net.Dialer
	useBind bool
}

func (r *systemDialer) dialAddress(ctx context.Context, network string, addr netip.Addr, port uint16) (net.Conn, error) {
	addrPort := netip.AddrPortFrom(addr.Unmap(), port)
	if !r.useBind {
		return r.dialer.DialContext(ctx, network, addrPort.String())
	}
	if r.useInterface {
		return r.interfaceDialer.DialContext(ctx, network, addrPort.String())
	}

	hostAddress := addrPort.Addr()

	var (
		dialer        net.Dialer
		addressIsIpv4 = internal.IsIPv4(hostAddress)
		networkIsUdp  = strings.HasPrefix(network, "udp")
		networkIsL3   = !slices.Contains(internal.L4NetworkList, network)
	)
	if addressIsIpv4 {
		switch {
		case networkIsL3:
			dialer = r.v4L3Dialer
		case networkIsUdp:
			dialer = r.v4UdpDialer
		default:
			dialer = r.v4TcpDialer
		}
	} else {
		switch {
		case networkIsL3:
			dialer = r.v6L3Dialer
		case networkIsUdp:
			dialer = r.v6UdpDialer
		default:
			dialer = r.v6TcpDialer
		}
	}
	return dialer.DialContext(ctx, network, addrPort.String())
}

func (r *systemDialer) DialParallel(ctx context.Context, addresses []netip.Addr, port uint16) (net.Conn, error) {
	if len(addresses) > 1 && r.fallbackDelay >= 0 {
		return DialParallel(ctx, r, addresses, port, r.dialStrategy == DialPreferIPv6, r.fallbackDelay)
	}

	return DialSerial(ctx, r, "tcp", addresses, port)
}

func (r *systemDialer) DialAddrPort(ctx context.Context, network string, addr netip.Addr, port uint16) (net.Conn, error) {
	return r.dialAddress(ctx, network, addr, port)
}

func (r *systemDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	if len(address) == 0 {
		return nil, ErrNoAddressToDialer
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
		if !slices.Contains(internal.L4NetworkList, network) {
			return nil, fmt.Errorf("address not resolved with network %s: %s", network, host)
		}
		addresses, err = internal.LocalLookup(ctx, host,
			network != "udp6" && network != "tcp6" && r.dialStrategy != DialOnlyIPv6,
			network != "udp4" && network != "tcp4" && r.dialStrategy != DialOnlyIPv4)
		if err != nil {
			return nil, fmt.Errorf("parse address: %s: lookup %s: %w", address, host, err)
		}
	} else {
		var addr netip.Addr
		if addr, err = netip.ParseAddr(host); err != nil {
			return nil, fmt.Errorf("parse address %s: %w", address, err)
		}
		if internal.IsIPv4(addr) && r.dialStrategy != DialOnlyIPv6 ||
			internal.IsIPv6(addr) && r.dialStrategy != DialOnlyIPv4 {
			return nil, fmt.Errorf("dialStrategy=%s,addr=%s: %w",
				r.dialStrategy, addr, ErrNoAddressToDialer)
		}
		addresses = []netip.Addr{addr}
	}

	if len(addresses) > 1 && r.fallbackDelay >= 0 && network == "tcp" {
		return r.DialParallel(ctx, addresses, uint16(hostPort))
	}

	return DialSerial(ctx, r, network, addresses, uint16(hostPort))
}

func NewDialerWithOption(opt ...DialerOption) Dialer {
	d := &systemDialer{
		dialer: net.Dialer{},
		happyEyeballConf: happyEyeballConf{
			fallbackDelay: internal.DefaultHappyEyeballFallbackDelay,
			dialStrategy:  DialPreferIPv6,
		},
	}

	var delayedOption []DialerOption
	for i := 0; i < len(opt); i++ {
		o := opt[i]
		if o.RequireDialer() {
			delayedOption = append(delayedOption, o)
			continue
		}
		o.Apply(d)
	}

	for i := 0; i < len(delayedOption); i++ {
		o := delayedOption[i]
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

func (ds *DialStrategy) UnmarshalYAML(b []byte) error {
	unqoted := mt.UnquoteString(string(b))
	if ss, ok := DialStrategyFromString(unqoted); ok {
		*ds = ss
	}
	return fmt.Errorf("unknown dialer strategy: `%s`", unqoted)
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
