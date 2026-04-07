package netxx

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/duakc/lightddns/infra/netxx/dnstransport"
)

var (
	DefaultResolveClient = NewDefaultDNSResolver()
)

type Dialer interface {
	DialContext(ctx context.Context, network string, address string) (net.Conn, error)
}

type bindAddressDialer struct {
	inet4Address netip.Addr
	inet6Address netip.Addr

	others Dialer

	tcpDialer4 Dialer
	tcpDialer6 Dialer

	udpDialer4 Dialer
	udpDialer6 Dialer
}

func (b *bindAddressDialer) init(d net.Dialer) {
	if addr := b.inet4Address; addr.IsValid() {
		tcpDialer, udpDialer := d, d
		tcpDialer.LocalAddr, udpDialer.LocalAddr = &net.TCPAddr{IP: addr.AsSlice()}, &net.UDPAddr{IP: addr.AsSlice()}
		b.tcpDialer4, b.udpDialer4 = &tcpDialer, &udpDialer
	} else {
		b.tcpDialer4, b.udpDialer4 = &d, &d
	}
	if addr := b.inet6Address; addr.IsValid() {
		tcpDialer, udpDialer := d, d
		tcpDialer.LocalAddr, udpDialer.LocalAddr = &net.TCPAddr{IP: addr.AsSlice()}, &net.UDPAddr{IP: addr.AsSlice()}
		b.tcpDialer6, b.udpDialer6 = &tcpDialer, &udpDialer
	} else {
		b.tcpDialer6, b.udpDialer6 = &d, &d
	}
	b.others = &d
}

func (b *bindAddressDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	if !strings.HasPrefix(network, "udp") && strings.HasPrefix(network, "tcp") {
		return b.others.DialContext(ctx, network, address)
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("SplitHostPort: %w", err)
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return nil, fmt.Errorf("%s not a valid ip address: %w", address, err)
	}
	var (
		v4Dialer Dialer = b.tcpDialer4
		v6Dialer Dialer = b.tcpDialer6
	)
	if strings.HasPrefix(network, "udp") {
		v4Dialer = b.udpDialer4
		v4Dialer = b.udpDialer6
	}
	if IsIPv4(addr) {
		return v4Dialer.DialContext(ctx, network, address)
	}
	return v6Dialer.DialContext(ctx, network, address)
}

type defaultDialer struct {
	dialer net.Dialer

	resolver  ResolveClient
	transport dnstransport.DNSTransport

	fallbackDelay time.Duration
	dialStrategy  DialStrategy

	bindDialer *bindAddressDialer
}

func (d *defaultDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	var dialer Dialer = &d.dialer
	if d.bindDialer != nil {
		dialer = d.bindDialer
	}

	if !(strings.HasPrefix(network, "udp") || strings.HasPrefix(network, "tcp")) {
		return dialer.DialContext(ctx, network, address)
	}
	var resolverStrategy uint = ResolveAsis
	if network == "udp6" || network == "tcp6" || d.dialStrategy == DialOnlyIPv6 {
		resolverStrategy = ResolveIPv6
	} else if network == "udp4" || network == "tcp4" || d.dialStrategy == DialPreferIPv4 {
		resolverStrategy = ResolveIPv4
	}

	host, portString, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("SplitHostPort: %w", err)
	}
	if !IsDomainName(host) {
		return dialer.DialContext(ctx, network, address)
	}
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("ParsePort: %w", err)
	}
	addresses, err := d.resolver.Lookup(ctx, d.transport, host, ResolveStrategy(resolverStrategy))
	if err != nil {
		return nil, fmt.Errorf("resolve failed: %w", err)
	}

	return DialParallel(ctx, dialer, network, addresses, uint16(port),
		d.dialStrategy == DialPreferIPv6 || d.dialStrategy == DialOnlyIPv6,
		defaultFallbackDelay)
}

func NewDialerWithOption(opt ...DialerOption) Dialer {
	d := &defaultDialer{
		dialer:        net.Dialer{},
		resolver:      DefaultResolveClient,
		transport:     &dnstransport.SystemTransport{},
		fallbackDelay: defaultFallbackDelay,
		dialStrategy:  DialPreferIPv6,
	}

	for i := 0; i < len(opt); i++ {
		o := opt[i]
		o.Apply(d)
	}

	if d.bindDialer != nil {
		d.bindDialer.init(d.dialer)
	}

	return d
}
