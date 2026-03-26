package ctxservice

import "context"

type Registry interface {
	Store(k, v any)
	Load(k any) any
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

func NewRegistry(ctx context.Context, r Registry) context.Context {
	ctx = context.WithValue(ctx, registryKey{}, r)
	return ctx
}

func RegistryFromContext(ctx context.Context) Registry {
	registry := ctx.Value(registryKey{})
	return registry.(Registry)
}
