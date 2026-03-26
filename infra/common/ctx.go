package common

import (
	"context"
	"time"
)

func ContextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	now := time.Now()
	if deadline, isDeadline := ctx.Deadline(); !isDeadline ||
		deadline.Before(now) || deadline.Sub(now) > timeout {
		return context.WithTimeout(ctx, timeout)
	}

	return ctx, func() {}
}
