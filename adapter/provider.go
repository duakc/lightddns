package adapter

import "net/netip"

type Provider interface {
	Type() string
	Name() string
	Diff(addr netip.Addr) (bool, error)
	Update(addr netip.Addr) error
}
