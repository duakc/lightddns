package ddnsmetric

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricProviderRequestTotal           = "request_total"
	MetricProviderRequestFailureTotal    = "request_failure_total"
	MetricProviderRequestDurationSeconds = "request_duration_seconds"

	MetricProviderLabelName      = "name"
	MetricProviderLabelOperation = "operation"
)

type providerLeaf struct {
	leaf
}

var ProviderLeaf providerLeaf

func (providerLeaf) RequestTotal(f Factory, name, op string) prometheus.Counter {
	return f.CounterVec(MetricProviderRequestTotal,
		"Total provider API requests.",
		[]string{MetricProviderLabelName, MetricProviderLabelOperation}).
		With(name, op)
}

func (providerLeaf) RequestFailure(f Factory, name, op string) prometheus.Counter {
	return f.CounterVec(MetricProviderRequestFailureTotal,
		"Failed provider API requests.",
		[]string{MetricProviderLabelName, MetricProviderLabelOperation}).
		With(name, op)
}

func (providerLeaf) RequestDuration(f Factory, name, op string, buckets []float64) prometheus.Observer {
	return f.HistogramVec(MetricProviderRequestDurationSeconds,
		"Provider API request duration.",
		[]string{MetricProviderLabelName, MetricProviderLabelOperation},
		buckets).
		With(name, op)
}
