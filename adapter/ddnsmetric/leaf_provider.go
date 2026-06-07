package ddnsmetric

import (
	"context"

	"github.com/duakc/lightddns/infra/metrics"

	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricProviderRequestTotal           = "request_total"
	MetricProviderRequestFailureTotal    = "request_failure_total"
	MetricProviderRequestDurationSeconds = "request_duration_seconds"
)

const (
	LabelProviderName = "name"
	LabelProviderOp   = "operation"
)

type ProviderFactory interface {
	services.ContextInjector
	metrics.Factory

	RequestTotal(name, op string) prometheus.Counter
	RequestFailure(name, op string) prometheus.Counter
	RequestDuration(name, op string, buckets []float64) prometheus.Observer
}

func NewProviderFactory(factory metrics.Factory) ProviderFactory {
	if sf, isSf := factory.(ProviderFactory); isSf {
		return sf
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

func (l *defaultProviderFactory) RequestTotal(name, op string) prometheus.Counter {
	return l.CounterVec(MetricProviderRequestTotal,
		"Total provider API requests.",
		[]string{LabelProviderName, LabelProviderOp}).
		With(name, op)
}

func (l *defaultProviderFactory) RequestFailure(name, op string) prometheus.Counter {
	return l.CounterVec(MetricProviderRequestFailureTotal,
		"Failed provider API requests.",
		[]string{LabelProviderName, LabelProviderOp}).
		With(name, op)
}

func (l *defaultProviderFactory) RequestDuration(name, op string, buckets []float64) prometheus.Observer {
	return l.HistogramVec(MetricProviderRequestDurationSeconds,
		"Provider API request duration.",
		[]string{LabelProviderName, LabelProviderOp},
		buckets).
		With(name, op)
}
