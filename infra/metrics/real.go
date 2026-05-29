package metrics

import (
	"context"
	"sync"

	"github.com/duakc/mt/services"

	"github.com/prometheus/client_golang/prometheus"
)

var _ Registry = (*defaultRegistry)(nil)

type defaultRegistry struct {
	reg *prometheus.Registry

	mu         sync.Mutex
	counters   map[string]CounterVec
	gauges     map[string]GaugeVec
	histograms map[string]HistogramVec
	summaries  map[string]SummaryVec
}

func newDefaultRegistry() Registry {
	return &defaultRegistry{
		reg:        prometheus.NewRegistry(),
		counters:   make(map[string]CounterVec),
		gauges:     make(map[string]GaugeVec),
		histograms: make(map[string]HistogramVec),
		summaries:  make(map[string]SummaryVec),
	}
}

func (r *defaultRegistry) ContextInject(ctx context.Context) context.Context {
	return services.InjectMe[Registry](ctx, r)
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
	if existing, ok := r.counters[opt.Name]; ok {
		return existing
	}
	vec := prometheus.NewCounterVec(opt, labels)
	r.reg.MustRegister(vec)
	c := &counterVec{vec: vec}
	r.counters[opt.Name] = c
	return c
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
	if existing, ok := r.gauges[opt.Name]; ok {
		return existing
	}
	vec := prometheus.NewGaugeVec(opt, labels)
	r.reg.MustRegister(vec)
	g := &gaugeVec{vec: vec}
	r.gauges[opt.Name] = g
	return g
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
	if existing, ok := r.histograms[opt.Name]; ok {
		return existing
	}
	if len(opt.Buckets) == 0 {
		opt.Buckets = prometheus.DefBuckets
	}
	vec := prometheus.NewHistogramVec(opt, labels)
	r.reg.MustRegister(vec)
	h := &histogramVec{vec: vec}
	r.histograms[opt.Name] = h
	return h
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
	if existing, ok := r.summaries[opt.Name]; ok {
		return existing
	}
	vec := prometheus.NewSummaryVec(opt, labels)
	r.reg.MustRegister(vec)
	s := &summaryVec{vec: vec}
	r.summaries[opt.Name] = s
	return s
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
