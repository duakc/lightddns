package metricx

import (
	"context"

	"github.com/duakc/lightddns/infra/metrics"

	"github.com/duakc/mt/services"
)

type ServiceFactory interface {
	services.ContextInjector
	metrics.Factory
}

func NewServiceFactory(factory metrics.Factory) ServiceFactory {
	if sf, isSf := factory.(*defaultServiceFactory); isSf {
		return sf
	}
	return &defaultServiceFactory{
		Factory: metrics.NewNameFactory(factory, Namespace, SubsystemService),
	}
}

var _ ServiceFactory = (*defaultServiceFactory)(nil)

type defaultServiceFactory struct {
	metrics.Factory
}

func (l *defaultServiceFactory) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[ServiceFactory](ctx, l)
}
