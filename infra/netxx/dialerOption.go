package netxx

import (
	"net"
	"net/netip"
	"time"

	"github.com/duakc/lightddns/infra/netxx/control"
	"github.com/duakc/lightddns/infra/netxx/dnstransport"
)

type DialerOption interface {
	Apply(d *defaultDialer)
}
type FuncDialerOption func(d *defaultDialer)

func (fn FuncDialerOption) Apply(d *defaultDialer) {
	fn(d)
}

func DialerOptionWithTransport(transport dnstransport.DNSTransport) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.transport = transport
	})
}

func DialerOptionWithDialStrategy(strategy DialStrategy) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.dialStrategy = strategy
	})
}

func DialerOptionWithDialer(dialer net.Dialer) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.dialer = dialer
	})
}

func DialerOptionWithFallbackDelay(delay time.Duration) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.fallbackDelay = delay
	})
}

func DialerOptionWithResolver(resolver ResolveClient) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.resolver = resolver
	})
}

func DialerOptionWithBindInterfaceName(interfaceName string) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.dialer.Control = control.Append(d.dialer.Control,
			control.BindToInterface(control.NewDefaultInterfaceFinder(),
				interfaceName, -1))
	})
}

func DialerOptionWithBindInterfaceIndex(interfaceIndex int) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.dialer.Control = control.Append(d.dialer.Control,
			control.BindToInterface(control.NewDefaultInterfaceFinder(),
				"", interfaceIndex))
	})
}

func DialerOptionBindAddress4(addr netip.Addr) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		if d.bindDialer == nil {
			d.bindDialer = &bindAddressDialer{}
		}
		d.bindDialer.inet4Address = addr
	})
}

func DialerOptionBindAddress6(addr netip.Addr) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		if d.bindDialer == nil {
			d.bindDialer = &bindAddressDialer{}
		}
		d.bindDialer.inet6Address = addr
	})
}

func DialerOptionFwmark(fwmark uint32) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.dialer.Control = control.Append(d.dialer.Control, control.RoutingMark(fwmark))
	})
}
