package metricx

import (
	"context"

	"github.com/duakc/lightddns/infra/metrics"

	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricProviderOperationTotal           = "operation_total"
	MetricProviderOperationFailureTotal    = "operation_failure_total"
	MetricProviderOperationDurationSeconds = "operation_duration_seconds"
)

const (
	LabelProviderName = "name"
	LabelProviderType = "type"
	LabelProviderOp   = "operation"
)

type ProviderFactory interface {
	services.ContextInjector
	metrics.Factory

	OperationTotal(name, providerType, op string) prometheus.Counter
	OperationFailure(name, providerType, op string) prometheus.Counter
	OperationDuration(name, providerType, op string, buckets []float64) prometheus.Observer
}

func NewProviderFactory(factory metrics.Factory) ProviderFactory {
	if providerFactory, ok := factory.(ProviderFactory); ok {
		return providerFactory
	}
	return &defaultProviderFactory{
		Factory: metrics.NewNameFactory(factory, Namespace, SubsystemProvider),
	}
}

var _ ProviderFactory = (*defaultProviderFactory)(nil)

type defaultProviderFactory struct {
	metrics.Factory
}

func (l *defaultProviderFactory) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[ProviderFactory](ctx, l)
}

func (l *defaultProviderFactory) OperationTotal(name, providerType, op string) prometheus.Counter {
	return l.CounterVec(MetricProviderOperationTotal,
		"Total provider operations.",
		[]string{LabelProviderName, LabelProviderType, LabelProviderOp}).
		With(name, providerType, op)
}

func (l *defaultProviderFactory) OperationFailure(name, providerType, op string) prometheus.Counter {
	return l.CounterVec(MetricProviderOperationFailureTotal,
		"Failed provider operations.",
		[]string{LabelProviderName, LabelProviderType, LabelProviderOp}).
		With(name, providerType, op)
}

func (l *defaultProviderFactory) OperationDuration(name, providerType, op string, buckets []float64) prometheus.Observer {
	return l.HistogramVec(MetricProviderOperationDurationSeconds,
		"Provider operation duration.",
		[]string{LabelProviderName, LabelProviderType, LabelProviderOp},
		buckets).
		With(name, providerType, op)
}
