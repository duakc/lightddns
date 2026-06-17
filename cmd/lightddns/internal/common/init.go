package common

import "context"

var globalContext context.Context

func StoreContext(ctx context.Context) {
	globalContext = ctx
}

func Context() context.Context {
	if globalContext == nil {
		globalContext = context.Background()
	}
	return globalContext
}
