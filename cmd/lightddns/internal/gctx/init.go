package gctx

import (
	"context"

	"github.com/duakc/lightddns/infra/gos"

	"github.com/duakc/mt/services"
)

func Context() (context.Context, context.CancelFunc) {
	ctx, cancel := gos.InterruptSignalContext(context.Background())
	ctx = services.NewRegistry(ctx, services.NewDefaultRegistry())
	return ctx, cancel
}
