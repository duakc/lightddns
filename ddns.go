package lightddns

import (
	"context"

	"github.com/duakc/lightddns/infra/ctxservice"
	"github.com/duakc/lightddns/options"
	"go.uber.org/zap"
)

type LightDDNS struct {
	logger *zap.Logger
}

func New(ctx context.Context, opt options.Options) (*LightDDNS, error) {
	ctx = ctxservice.NewRegistry(ctx, ctxservice.NewDefaultRegistry())

}
