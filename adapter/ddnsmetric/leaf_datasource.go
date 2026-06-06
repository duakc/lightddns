package ddnsmetric

import (
	"context"

	"github.com/duakc/lightddns/infra/metrics"
	"github.com/duakc/mt/services"
)

type DataSourceFactory interface {
	services.ContextInjector
	metrics.Factory
}

func NewDatasourceFactory(factory metrics.Factory) DataSourceFactory {
	if df, isDf := factory.(*defaultDatasourceFactory); isDf {
		return df
	}
	return &defaultDatasourceFactory{
		Factory: factory,
	}
}

var _ DataSourceFactory = (*defaultDatasourceFactory)(nil)

type defaultDatasourceFactory struct {
	metrics.Factory
}

func (l *defaultDatasourceFactory) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[DataSourceFactory](ctx, l)
}
