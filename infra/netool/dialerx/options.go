package dialerx

import (
	"net"
	"net/netip"
	"time"

	"github.com/duakc/lightddns/infra/netool/control"
)

type DialerOption interface {
	Apply(d *systemDialer)
	RequireDialer() bool
}
type FuncDialerOption func(d *systemDialer)

func (fn FuncDialerOption) Apply(d *systemDialer) {
	fn(d)
}

func (fn FuncDialerOption) RequireDialer() bool {
	return false
}

func WithDialStrategy(strategy DialStrategy) DialerOption {
	return FuncDialerOption(func(d *systemDialer) {
		d.dialStrategy = strategy
	})
}

func WithDialer(dialer net.Dialer) DialerOption {
	return FuncDialerOption(func(d *systemDialer) {
		d.dialer = dialer
	})
}

func WithFallbackDelay(delay time.Duration) DialerOption {
	return FuncDialerOption(func(d *systemDialer) {
		d.fallbackDelay = delay
	})
}

func WithFwmark(fwmark uint32) DialerOption {
	return FuncDialerOption(func(d *systemDialer) {
		d.dialer.Control = control.Append(d.dialer.Control, control.RoutingMark(fwmark))
	})
}

type bindOption struct {
	interfaceName  string
	interfaceIndex int

	address4 netip.Addr
	address6 netip.Addr
}

func (o *bindOption) RequireDialer() bool {
	return true
}

func (o *bindOption) Apply(d *systemDialer) {
	d.useBind = true
	if o.interfaceName != "" || o.interfaceIndex > 0 {
		d.bindDialer = newInterfaceBindDialer(d.dialer, o.interfaceName, o.interfaceIndex, control.NewDefaultInterfaceFinder())
	} else if o.address4.IsValid() || o.address6.IsValid() {
		d.bindDialer = newAddressesBindDialer(d.dialer, o.address4, o.address6)
	} else {
		d.useBind = false
	}
}

func WithBindInterfaceName(interfaceName string) DialerOption {
	return &bindOption{interfaceName: interfaceName}
}

func WithBindInterfaceIndex(interfaceIndex int) DialerOption {
	return &bindOption{interfaceIndex: interfaceIndex}
}

func WithBindAddress(addr4, addr6 netip.Addr) DialerOption {
	return &bindOption{address4: addr4, address6: addr6}
}
