package common

import (
	"context"
	"time"
)

func ContextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, isDeadline := ctx.Deadline(); !isDeadline {
		return context.WithTimeout(ctx, timeout)
	}

	return ctx, func() {}
}
