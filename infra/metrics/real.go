package metrics

import (
	"context"

	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/mt/debug"
	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var realLogger = zaplog.NewPackage("infra", "metrics")

type defaultRegistry struct {
	reg        *prometheus.Registry
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
}

func newDefaultRegistry() Registry {
	r := &defaultRegistry{
		reg:        prometheus.NewRegistry(),
		counters:   make(map[string]*prometheus.CounterVec, len(acquiredCounters)),
		gauges:     make(map[string]*prometheus.GaugeVec, len(acquiredGauges)),
		histograms: make(map[string]*prometheus.HistogramVec, len(acquiredHistograms)),
	}
	for _, d := range acquiredCounters {
		vec := prometheus.NewCounterVec(prometheus.CounterOpts{Name: d.name, Help: d.help}, d.labels)
		r.reg.MustRegister(vec)
		r.counters[d.name] = vec
	}
	for _, d := range acquiredGauges {
		vec := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: d.name, Help: d.help}, d.labels)
		r.reg.MustRegister(vec)
		r.gauges[d.name] = vec
	}
	for _, d := range acquiredHistograms {
		vec := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Name: d.name, Help: d.help, Buckets: d.buckets},
			d.labels)
		r.reg.MustRegister(vec)
		r.histograms[d.name] = vec
	}
	return r
}

func (r *defaultRegistry) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[Registry](ctx, r)
}

func (r *defaultRegistry) Counter(name string, labelValues ...string) prometheus.Counter {
	vec, ok := r.counters[name]
	if !ok {
		return missingCounter(name)
	}
	return vec.WithLabelValues(labelValues...)
}

func (r *defaultRegistry) Gauge(name string, labelValues ...string) prometheus.Gauge {
	vec, ok := r.gauges[name]
	if !ok {
		return missingGauge(name)
	}
	return vec.WithLabelValues(labelValues...)
}

func (r *defaultRegistry) Histogram(name string, labelValues ...string) prometheus.Observer {
	vec, ok := r.histograms[name]
	if !ok {
		return missingHistogram(name)
	}
	return vec.WithLabelValues(labelValues...)
}

func (r *defaultRegistry) Gatherer() prometheus.Gatherer { return r.reg }

func missingCounter(name string) prometheus.Counter {
	if debug.Enabled {
		panic("metrics: counter not registered: " + name)
	}
	realLogger.Warn("counter not registered", zap.String("name", name))
	return nopCounter{}
}

func missingGauge(name string) prometheus.Gauge {
	if debug.Enabled {
		panic("metrics: gauge not registered: " + name)
	}
	realLogger.Warn("gauge not registered", zap.String("name", name))
	return nopGauge{}
}

func missingHistogram(name string) prometheus.Observer {
	if debug.Enabled {
		panic("metrics: histogram not registered: " + name)
	}
	realLogger.Warn("histogram not registered", zap.String("name", name))
	return nopObserver{}
}
