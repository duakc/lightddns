package adapter

import (
	"context"
	"net/netip"
)

type Provider interface {
	managedType
	Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error)
	Update(ctx context.Context, domain string, ttl int, addr []netip.Addr) error
}

type ProviderManager = Manager[Provider]

var ProviderRegister = NewRegister[Provider]()
