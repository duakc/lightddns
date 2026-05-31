package zaplog

import (
	"context"

	"github.com/duakc/mt/services/container"

	"go.uber.org/zap"
)

const ContainerLoggerName = "zaplog.logger.stack"

type loggerStack struct {
	st []*zap.Logger
}

func From(c container.Container) *zap.Logger {
	stack, ok := container.LoadPtr[loggerStack](c, ContainerLoggerName)
	if !ok || stack == nil || len(stack.st) == 0 {
		return defaultLogger
	}
	return stack.st[len(stack.st)-1]
}

func With(c container.Container, logger *zap.Logger) {
	stack, ok := container.LoadPtr[loggerStack](c, ContainerLoggerName)
	if !ok {
		stack = &loggerStack{}
	}
	stack.st = append(stack.st, logger)
	container.StorePtr[loggerStack](c, ContainerLoggerName, stack)
}

func Kick(c container.Container) {
	stack, ok := container.LoadPtr[loggerStack](c, ContainerLoggerName)
	if !ok || stack == nil || len(stack.st) == 0 {
		return
	}
	stack.st[len(stack.st)-1] = nil
	stack.st = stack.st[:len(stack.st)-1]
	container.StorePtr[loggerStack](c, ContainerLoggerName, stack)
}

func FromContext(ctx context.Context) *zap.Logger {
	stack, ok := container.LoadPtrContext[loggerStack](ctx, ContainerLoggerName)
	if !ok || stack == nil || len(stack.st) == 0 {
		return defaultLogger
	}
	return stack.st[len(stack.st)-1]
}

func WithContext(ctx context.Context, logger *zap.Logger) {
	stack, ok := container.LoadPtrContext[loggerStack](ctx, ContainerLoggerName)
	if !ok {
		stack = &loggerStack{}
	}
	stack.st = append(stack.st, logger)
	container.StorePtrContext[loggerStack](ctx, ContainerLoggerName, stack)
}

func KickContext(ctx context.Context) {
	stack, ok := container.LoadPtrContext[loggerStack](ctx, ContainerLoggerName)
	if !ok || stack == nil || len(stack.st) == 0 {
		return
	}
	stack.st[len(stack.st)-1] = nil
	stack.st = stack.st[:len(stack.st)-1]
	container.StorePtrContext[loggerStack](ctx, ContainerLoggerName, stack)
}

func FromOrPackage(ctx context.Context, pkg ...string) *zap.Logger {
	if c, ok := container.FromContext(ctx); ok {
		if stack, sok := container.LoadPtr[loggerStack](c, ContainerLoggerName); sok &&
			stack != nil && len(stack.st) > 0 {
			return stack.st[len(stack.st)-1]
		}
	}
	return NewPackage(pkg...)
}

func FromOrDefault(ctx context.Context, dft *zap.Logger) *zap.Logger {
	if c, ok := container.FromContext(ctx); ok {
		if stack, sok := container.LoadPtr[loggerStack](c, ContainerLoggerName); sok &&
			stack != nil && len(stack.st) > 0 {
			return stack.st[len(stack.st)-1]
		}
	}
	return dft
}
