package ctxservice

import "context"

type Registry interface {
	Store(k, v any)
	Load(k any) any

	Clear()
}

type registryKey struct{}

func Lookup[V any](ctx context.Context, key any) V {
	registry := RegistryFromContext(ctx)
	return registry.Load(key).(V)
}

func Store[K comparable, V any](ctx context.Context, key K, value V) {
	registry := RegistryFromContext(ctx)
	registry.Store(key, value)
}

func NewRegistry(ctx context.Context) (context.Context, Registry) {
	r := ctx.Value(registryKey{})
	if r == nil {
		reg := newDefaultRegistry()
		ctx = context.WithValue(ctx, registryKey{}, r)
		return ctx, reg
	}
	reg := r.(Registry)
	reg.Clear()
	return ctx, reg
}

func RegistryFromContext(ctx context.Context) Registry {
	registry := ctx.Value(registryKey{})
	return registry.(Registry)
}
