package metrics

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sync"

	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus"
)

var _ Registry = (*defaultRegistry)(nil)

type defaultRegistry struct {
	reg *prometheus.Registry

	mu         sync.Mutex
	counters   map[string]registeredMetric[CounterVec]
	gauges     map[string]registeredMetric[GaugeVec]
	histograms map[string]registeredMetric[HistogramVec]
	summaries  map[string]registeredMetric[SummaryVec]
}

type registeredMetric[T any] struct {
	metric T
	opts   any
	labels []string
}

func newDefaultRegistry() Registry {
	return &defaultRegistry{
		reg:        prometheus.NewRegistry(),
		counters:   make(map[string]registeredMetric[CounterVec]),
		gauges:     make(map[string]registeredMetric[GaugeVec]),
		histograms: make(map[string]registeredMetric[HistogramVec]),
		summaries:  make(map[string]registeredMetric[SummaryVec]),
	}
}

func (r *defaultRegistry) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[Registry](ctx, r)
}

func metricKey(namespace, subsystem, name string) string {
	return prometheus.BuildFQName(namespace, subsystem, name)
}

func registeredOrPanic[T any](registered registeredMetric[T], opts any, labels []string, name string) T {
	if !reflect.DeepEqual(registered.opts, opts) || !slices.Equal(registered.labels, labels) {
		panic(fmt.Sprintf("metrics: %s registered with a different schema", name))
	}
	return registered.metric
}

func (r *defaultRegistry) CounterVec(name, help string, labels []string) CounterVec {
	if name == "" {
		panic("CounterVec must have a name")
	}
	return r.CounterVecVerbose(prometheus.CounterOpts{Name: name, Help: help}, labels)
}

func (r *defaultRegistry) CounterVecVerbose(opt prometheus.CounterOpts, labels []string) CounterVec {
	r.mu.Lock()
	defer r.mu.Unlock()
	if opt.Name == "" {
		panic("CounterVecVerbose must have a name")
	}
	key := metricKey(opt.Namespace, opt.Subsystem, opt.Name)
	if existing, ok := r.counters[key]; ok {
		return registeredOrPanic(existing, opt, labels, key)
	}
	vec := prometheus.NewCounterVec(opt, labels)
	r.reg.MustRegister(vec)
	counter := &counterVec{vec: vec}
	r.counters[key] = registeredMetric[CounterVec]{
		metric: counter,
		opts:   opt,
		labels: slices.Clone(labels),
	}
	return counter
}

func (r *defaultRegistry) GaugeVec(name, help string, labels []string) GaugeVec {
	if name == "" {
		panic("GaugeVec must have a name")
	}
	return r.GaugeVecVerbose(prometheus.GaugeOpts{Name: name, Help: help}, labels)
}

func (r *defaultRegistry) GaugeVecVerbose(opt prometheus.GaugeOpts, labels []string) GaugeVec {
	r.mu.Lock()
	defer r.mu.Unlock()
	if opt.Name == "" {
		panic("GaugeVecVerbose must have a name")
	}
	key := metricKey(opt.Namespace, opt.Subsystem, opt.Name)
	if existing, ok := r.gauges[key]; ok {
		return registeredOrPanic(existing, opt, labels, key)
	}
	vec := prometheus.NewGaugeVec(opt, labels)
	r.reg.MustRegister(vec)
	gauge := &gaugeVec{vec: vec}
	r.gauges[key] = registeredMetric[GaugeVec]{
		metric: gauge,
		opts:   opt,
		labels: slices.Clone(labels),
	}
	return gauge
}

func (r *defaultRegistry) HistogramVec(name, help string, labels []string, buckets []float64) HistogramVec {
	if name == "" {
		panic("HistogramVec must have a name")
	}
	return r.HistogramVecVerbose(
		prometheus.HistogramOpts{Name: name, Help: help, Buckets: buckets},
		labels)
}

func (r *defaultRegistry) HistogramVecVerbose(opt prometheus.HistogramOpts, labels []string) HistogramVec {
	r.mu.Lock()
	defer r.mu.Unlock()
	if opt.Name == "" {
		panic("HistogramVecVerbose must have a name")
	}
	if len(opt.Buckets) == 0 {
		opt.Buckets = prometheus.DefBuckets
	}
	key := metricKey(opt.Namespace, opt.Subsystem, opt.Name)
	if existing, ok := r.histograms[key]; ok {
		return registeredOrPanic(existing, opt, labels, key)
	}
	vec := prometheus.NewHistogramVec(opt, labels)
	r.reg.MustRegister(vec)
	histogram := &histogramVec{vec: vec}
	r.histograms[key] = registeredMetric[HistogramVec]{
		metric: histogram,
		opts:   opt,
		labels: slices.Clone(labels),
	}
	return histogram
}

func (r *defaultRegistry) SummaryVec(name, help string, labels []string, objectives map[float64]float64) SummaryVec {
	if name == "" {
		panic("SummaryVec must have a name")
	}
	return r.SummaryVecVerbose(
		prometheus.SummaryOpts{Name: name, Help: help, Objectives: objectives},
		labels)
}

func (r *defaultRegistry) SummaryVecVerbose(opt prometheus.SummaryOpts, labels []string) SummaryVec {
	r.mu.Lock()
	defer r.mu.Unlock()
	if opt.Name == "" {
		panic("SummaryVecVerbose must have a name")
	}
	key := metricKey(opt.Namespace, opt.Subsystem, opt.Name)
	if existing, ok := r.summaries[key]; ok {
		return registeredOrPanic(existing, opt, labels, key)
	}
	vec := prometheus.NewSummaryVec(opt, labels)
	r.reg.MustRegister(vec)
	summary := &summaryVec{vec: vec}
	r.summaries[key] = registeredMetric[SummaryVec]{
		metric: summary,
		opts:   opt,
		labels: slices.Clone(labels),
	}
	return summary
}

func (r *defaultRegistry) Gatherer() prometheus.Gatherer { return r.reg }

type counterVec struct{ vec *prometheus.CounterVec }

func (c *counterVec) With(labelValues ...string) prometheus.Counter {
	return c.vec.WithLabelValues(labelValues...)
}

type gaugeVec struct{ vec *prometheus.GaugeVec }

func (g *gaugeVec) With(labelValues ...string) prometheus.Gauge {
	return g.vec.WithLabelValues(labelValues...)
}

type histogramVec struct{ vec *prometheus.HistogramVec }

func (h *histogramVec) With(labelValues ...string) prometheus.Observer {
	return h.vec.WithLabelValues(labelValues...)
}

type summaryVec struct{ vec *prometheus.SummaryVec }

func (s *summaryVec) With(labelValues ...string) prometheus.Observer {
	return s.vec.WithLabelValues(labelValues...)
}
