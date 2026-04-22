package adapter

import (
	"context"
	"net/netip"
)

type Provider interface {
	managedType
	Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error)
	Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) error
}

type (
	ProviderManager    = DefaultManager[Provider]
	ProviderManagerKey struct{}
)

var ProviderRegister = NewRegister[Provider]()
