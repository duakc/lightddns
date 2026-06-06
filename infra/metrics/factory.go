package metrics

import "github.com/prometheus/client_golang/prometheus"

type Factory interface {
	CounterVec(name, help string, labels []string) CounterVec
	CounterVecVerbose(opt prometheus.CounterOpts, labels []string) CounterVec

	GaugeVec(name, help string, labels []string) GaugeVec
	GaugeVecVerbose(opt prometheus.GaugeOpts, labels []string) GaugeVec

	HistogramVec(name, help string, labels []string, buckets []float64) HistogramVec
	HistogramVecVerbose(opt prometheus.HistogramOpts, labels []string) HistogramVec

	SummaryVec(name, help string, labels []string, objectives map[float64]float64) SummaryVec
	SummaryVecVerbose(opt prometheus.SummaryOpts, labels []string) SummaryVec
}

func NewNameFactory(reg Factory, namespace, subsystem string) Factory {
	return &factory{registry: reg, namespace: namespace, subsystem: subsystem}
}

type factory struct {
	registry  Factory
	namespace string
	subsystem string
}

func (o *factory) CounterVec(name, help string, labels []string) CounterVec {
	return o.CounterVecVerbose(prometheus.CounterOpts{Name: name, Help: help}, labels)
}

func (o *factory) CounterVecVerbose(opt prometheus.CounterOpts, labels []string) CounterVec {
	opt.Namespace, opt.Subsystem = o.namespace, o.subsystem
	return o.registry.CounterVecVerbose(opt, labels)
}

func (o *factory) GaugeVec(name, help string, labels []string) GaugeVec {
	return o.GaugeVecVerbose(prometheus.GaugeOpts{Name: name, Help: help}, labels)
}

func (o *factory) GaugeVecVerbose(opt prometheus.GaugeOpts, labels []string) GaugeVec {
	opt.Namespace, opt.Subsystem = o.namespace, o.subsystem
	return o.registry.GaugeVecVerbose(opt, labels)
}

func (o *factory) HistogramVec(name, help string, labels []string, buckets []float64) HistogramVec {
	return o.HistogramVecVerbose(
		prometheus.HistogramOpts{Name: name, Help: help, Buckets: buckets},
		labels)
}

func (o *factory) HistogramVecVerbose(opt prometheus.HistogramOpts, labels []string) HistogramVec {
	opt.Namespace, opt.Subsystem = o.namespace, o.subsystem
	return o.registry.HistogramVecVerbose(opt, labels)
}

func (o *factory) SummaryVec(name, help string, labels []string, objectives map[float64]float64) SummaryVec {
	return o.SummaryVecVerbose(
		prometheus.SummaryOpts{Name: name, Help: help, Objectives: objectives},
		labels)
}

func (o *factory) SummaryVecVerbose(opt prometheus.SummaryOpts, labels []string) SummaryVec {
	opt.Namespace, opt.Subsystem = o.namespace, o.subsystem
	return o.registry.SummaryVecVerbose(opt, labels)
}
