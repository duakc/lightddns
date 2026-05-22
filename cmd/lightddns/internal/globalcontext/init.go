package globalcontext

import "context"

var globalContext context.Context

func Store(ctx context.Context) {
	globalContext = ctx
}

func Load() context.Context {
	if globalContext == nil {
		globalContext = context.Background()
	}
	return globalContext
}
