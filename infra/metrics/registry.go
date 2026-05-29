package metrics

import (
	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus"
)

type Registry interface {
	services.ContextInjector

	// Counter
	CounterVec(name, help string, labels []string) CounterVec
	CounterVecVerbose(opt prometheus.CounterOpts, labels []string) CounterVec

	// Gauge
	GaugeVec(name, help string, labels []string) GaugeVec
	GaugeVecVerbose(opt prometheus.GaugeOpts, labels []string) GaugeVec

	// Histogram
	HistogramVec(name, help string, labels []string, buckets []float64) HistogramVec
	HistogramVecVerbose(opt prometheus.HistogramOpts, labels []string) HistogramVec

	// Summary
	SummaryVec(name, help string, labels []string, objectives map[float64]float64) SummaryVec
	SummaryVecVerbose(opt prometheus.SummaryOpts, labels []string) SummaryVec

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
