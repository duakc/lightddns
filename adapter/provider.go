package adapter

import (
	"context"
	"fmt"
	"net/netip"
)

type Provider interface {
	managedType
	Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error)
	Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) error
}

type (
	ProviderManager = Manager[Provider]
)

var ProviderRegister = NewRegister[Provider]()

type ProviderNotFoundError struct {
	*ManagedNotFoundError
}

func (e *ProviderNotFoundError) Error() string {
	return fmt.Sprintf("provider: %s", e.ManagedNotFoundError.Error())
}
