package adapter

import (
	"context"
	"fmt"
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

func LookupAllProvider(manager ProviderManager, names []string) ([]Provider, error) {
	all, err := manager.LookupAll(names)
	if err != nil {
		return nil, &ProviderNotFoundError{err}
	}
	return all, nil
}

type ProviderNotFoundError struct {
	Err error
}

func (e *ProviderNotFoundError) Error() string {
	return fmt.Sprintf("provider not found: %s", e.Err.Error())
}

func (e *ProviderNotFoundError) Unwrap() error {
	return e.Err
}
