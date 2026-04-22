package providers

import (
	"github.com/duakc/lightddns/options"

	"go.uber.org/zap"
)

func NewLogger(logger *zap.Logger, opt options.AbstractProviderOption) *zap.Logger {
	return logger.With(
		zap.String("type", "provider"),
		zap.String("provider_type", opt.Type)).
		Named(opt.Name)
}
