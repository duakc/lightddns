package zaplog

import (
	"context"

	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/container"

	"go.uber.org/zap"
)

type Factory interface {
	services.ContextInjector

	Logger(ctx context.Context) *zap.Logger
	Pass(ctx context.Context, logger *zap.Logger)
}

const ContainerLoggerName = "zaplog.logger"

var _ Factory = (*DefaultFactory)(nil)

type DefaultFactory struct {
	logger *zap.Logger
}

func NewFactory(logger *zap.Logger) *DefaultFactory {
	return &DefaultFactory{logger: DoNotPanic(logger)}
}

func FromContext(ctx context.Context) *zap.Logger {
	fa := services.Lookup[Factory](ctx)
	if fa == nil {
		return NOP
	}
	return fa.Logger(ctx)
}

func WithContext(ctx context.Context, logger *zap.Logger) {
	fa := services.Lookup[Factory](ctx)
	if fa == nil {
		return
	}
	fa.Pass(ctx, logger)
}

func (f *DefaultFactory) Logger(ctx context.Context) *zap.Logger {
	if f == nil {
		return NOP
	}
	parentLogger, ok := container.LoadPtrContext[zap.Logger](ctx, ContainerLoggerName)
	if ok {
		return parentLogger
	}

	return f.logger
}

func (f *DefaultFactory) Pass(ctx context.Context, logger *zap.Logger) {
	if logger == nil || ctx == nil || f == nil {
		return
	}

	container.StorePtrContext[zap.Logger](ctx, ContainerLoggerName, logger)
}

func (f *DefaultFactory) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[Factory](ctx, f)
}
