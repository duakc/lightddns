package metrics

import (
	"context"

	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var _ Registry = (*noopRegistry)(nil)

type noopRegistry struct{}

func newNoopRegistry() Registry { return noopRegistry{} }

func (noopRegistry) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[Registry](ctx, noopRegistry{})
}

func (noopRegistry) CounterVec(string, string, []string) CounterVec { return nopCounterVec{} }
func (noopRegistry) CounterVecVerbose(prometheus.CounterOpts, []string) CounterVec {
	return nopCounterVec{}
}

func (noopRegistry) GaugeVec(string, string, []string) GaugeVec { return nopGaugeVec{} }
func (noopRegistry) GaugeVecVerbose(prometheus.GaugeOpts, []string) GaugeVec {
	return nopGaugeVec{}
}

func (noopRegistry) HistogramVec(string, string, []string, []float64) HistogramVec {
	return nopHistogramVec{}
}

func (noopRegistry) HistogramVecVerbose(prometheus.HistogramOpts, []string) HistogramVec {
	return nopHistogramVec{}
}

func (noopRegistry) SummaryVec(string, string, []string, map[float64]float64) SummaryVec {
	return nopSummaryVec{}
}

func (noopRegistry) SummaryVecVerbose(prometheus.SummaryOpts, []string) SummaryVec {
	return nopSummaryVec{}
}

func (noopRegistry) Gatherer() prometheus.Gatherer { return prometheus.NewRegistry() }

type nopCounterVec struct{}

func (nopCounterVec) With(...string) prometheus.Counter { return nopCounter{} }

type nopGaugeVec struct{}

func (nopGaugeVec) With(...string) prometheus.Gauge { return nopGauge{} }

type nopHistogramVec struct{}

func (nopHistogramVec) With(...string) prometheus.Observer { return nopObserver{} }

type nopSummaryVec struct{}

func (nopSummaryVec) With(...string) prometheus.Observer { return nopObserver{} }

type nopCounter struct{}

func (nopCounter) Desc() *prometheus.Desc           { return nil }
func (nopCounter) Write(*dto.Metric) error          { return nil }
func (nopCounter) Describe(chan<- *prometheus.Desc) {}
func (nopCounter) Collect(chan<- prometheus.Metric) {}
func (nopCounter) Inc()                             {}
func (nopCounter) Add(float64)                      {}

type nopGauge struct{}

func (nopGauge) Desc() *prometheus.Desc           { return nil }
func (nopGauge) Write(*dto.Metric) error          { return nil }
func (nopGauge) Describe(chan<- *prometheus.Desc) {}
func (nopGauge) Collect(chan<- prometheus.Metric) {}
func (nopGauge) Set(float64)                      {}
func (nopGauge) Inc()                             {}
func (nopGauge) Dec()                             {}
func (nopGauge) Add(float64)                      {}
func (nopGauge) Sub(float64)                      {}
func (nopGauge) SetToCurrentTime()                {}

type nopObserver struct{}

func (nopObserver) Observe(float64) {}
