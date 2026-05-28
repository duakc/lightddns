package adapter

import (
	"context"
	"fmt"
	"net/netip"

	"go.uber.org/zap"
)

type Provider interface {
	managedType
	Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error)
	Update(ctx context.Context, domain string, ttl uint32, addr []netip.Addr) (changed bool, err error)
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

func CreatProviderLogger(logger *zap.Logger, provider Provider) *zap.Logger {
	return logger.With(
		zap.String("provider", provider.Name()),
		zap.String("provider_type", provider.Type())).
		Named("provider")
}
