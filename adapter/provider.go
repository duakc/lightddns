package adapter

import (
	"context"
	"net/netip"
)

type Provider interface {
	ManagedType
	Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error)
	Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (changed bool, err error)
}

type (
	ProviderManager = Manager[Provider]
)

var ProviderRegister = NewRegister[Provider]()
