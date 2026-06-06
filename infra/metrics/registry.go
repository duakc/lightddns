package metrics

import (
	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus"
)

type Registry interface {
	services.ContextInjector
	Factory

	Gatherer() prometheus.Gatherer
}

type CounterVec interface {
	With(labelValues ...string) prometheus.Counter
}

type GaugeVec interface {
	With(labelValues ...string) prometheus.Gauge
}

type HistogramVec interface {
	With(labelValues ...string) prometheus.Observer
}

type SummaryVec interface {
	With(labelValues ...string) prometheus.Observer
}

func New(enabled bool) Registry {
	if !enabled {
		return newNoopRegistry()
	}
	return newDefaultRegistry()
}
