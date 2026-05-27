package metrics

import (
	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus"
)

type Registry interface {
	services.ContextInjector

	Counter(name string, labelValues ...string) prometheus.Counter
	Gauge(name string, labelValues ...string) prometheus.Gauge
	Histogram(name string, labelValues ...string) prometheus.Observer

	Gatherer() prometheus.Gatherer
}

func New(enabled bool) Registry {
	if !enabled {
		return newNoopRegistry()
	}
	return newDefaultRegistry()
}
