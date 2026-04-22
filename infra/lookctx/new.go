package lookctx

import (
	"context"

	"github.com/duakc/mt"
)

type Registry interface {
	Store(k, v any)
	Load(k any) any
}

type registryKey struct{}

func Lookup[K comparable, V any](ctx context.Context) V {
	registry := RegistryFromContext(ctx)
	return registry.Load(mt.Zero[K]()).(V)
}

func Store[K comparable, V any](ctx context.Context, value V) {
	registry := RegistryFromContext(ctx)
	registry.Store(mt.Zero[K](), value)
}

func NewRegistry(ctx context.Context, r Registry) context.Context {
	return context.WithValue(ctx, registryKey{}, r)
}

func RegistryFromContext(ctx context.Context) Registry {
	registry := ctx.Value(registryKey{})
	return registry.(Registry)
}
