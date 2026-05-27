package metrics

import (
	"context"

	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type noopRegistry struct{}

func newNoopRegistry() Registry { return noopRegistry{} }

func (noopRegistry) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[Registry](ctx, noopRegistry{})
}

func (noopRegistry) Counter(string, ...string) prometheus.Counter    { return nopCounter{} }
func (noopRegistry) Gauge(string, ...string) prometheus.Gauge        { return nopGauge{} }
func (noopRegistry) Histogram(string, ...string) prometheus.Observer { return nopObserver{} }
func (noopRegistry) Gatherer() prometheus.Gatherer                   { return prometheus.NewRegistry() }

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
