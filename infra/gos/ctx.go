package gos

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func InterruptSignalContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, os.Interrupt, os.Kill,
		syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)
}
