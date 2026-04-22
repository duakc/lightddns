package dialerx

import (
	"net"
	"net/netip"
	"time"

	"github.com/duakc/lightddns/infra/netool/control"
)

type DialerOption interface {
	Apply(d *defaultDialer)
}
type FuncDialerOption func(d *defaultDialer)

func (fn FuncDialerOption) Apply(d *defaultDialer) {
	fn(d)
}

func WithDialStrategy(strategy DialStrategy) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.dialStrategy = strategy
	})
}

func WithDialer(dialer net.Dialer) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.dialer = dialer
	})
}

func WithFallbackDelay(delay time.Duration) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.fallbackDelay = delay
	})
}

// WithBindInterfaceName
// Warning: put this DialerOption behind other DialerOption.
func WithBindInterfaceName(interfaceName string) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.bind = newInterfaceBindDialer(d.dialer, interfaceName,
			0, control.NewDefaultInterfaceFinder())
	})
}

// WithBindInterfaceIndex
// Warning: put this DialerOption behind other DialerOption.
func WithBindInterfaceIndex(interfaceIndex int) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.bind = newInterfaceBindDialer(d.dialer, "",
			interfaceIndex, control.NewDefaultInterfaceFinder())

	})
}

// WithBindAddress
// Warning: put this DialerOption behind other DialerOption.
func WithBindAddress(addr4, addr6 netip.Addr) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.bind = newAddressesBindDialer(d.dialer, addr4, addr6)
	})
}

func WithFwmark(fwmark uint32) DialerOption {
	return FuncDialerOption(func(d *defaultDialer) {
		d.dialer.Control = control.Append(d.dialer.Control, control.RoutingMark(fwmark))
	})
}
