package ddnsmetric

import (
	"context"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/metrics"

	"github.com/duakc/mt/services/container"

	"github.com/prometheus/client_golang/prometheus"
)

type Factory interface {
	CounterVec(name, help string, labels []string) metrics.CounterVec
	CounterVecVerbose(opt prometheus.CounterOpts, labels []string) metrics.CounterVec

	GaugeVec(name, help string, labels []string) metrics.GaugeVec
	GaugeVecVerbose(opt prometheus.GaugeOpts, labels []string) metrics.GaugeVec

	HistogramVec(name, help string, labels []string, buckets []float64) metrics.HistogramVec
	HistogramVecVerbose(opt prometheus.HistogramOpts, labels []string) metrics.HistogramVec

	SummaryVec(name, help string, labels []string, objectives map[float64]float64) metrics.SummaryVec
	SummaryVecVerbose(opt prometheus.SummaryOpts, labels []string) metrics.SummaryVec
}

func NewFactory(reg metrics.Registry, namespace, subsystem string) Factory {
	return &factory{registry: reg, namespace: namespace, subsystem: subsystem}
}

type factory struct {
	registry  metrics.Registry
	namespace string
	subsystem string
}

func (o *factory) CounterVec(name, help string, labels []string) metrics.CounterVec {
	return o.CounterVecVerbose(prometheus.CounterOpts{Name: name, Help: help}, labels)
}

func (o *factory) CounterVecVerbose(opt prometheus.CounterOpts, labels []string) metrics.CounterVec {
	opt.Namespace, opt.Subsystem = o.namespace, o.subsystem
	return o.registry.CounterVecVerbose(opt, labels)
}

func (o *factory) GaugeVec(name, help string, labels []string) metrics.GaugeVec {
	return o.GaugeVecVerbose(prometheus.GaugeOpts{Name: name, Help: help}, labels)
}

func (o *factory) GaugeVecVerbose(opt prometheus.GaugeOpts, labels []string) metrics.GaugeVec {
	opt.Namespace, opt.Subsystem = o.namespace, o.subsystem
	return o.registry.GaugeVecVerbose(opt, labels)
}

func (o *factory) HistogramVec(name, help string, labels []string, buckets []float64) metrics.HistogramVec {
	return o.HistogramVecVerbose(
		prometheus.HistogramOpts{Name: name, Help: help, Buckets: buckets},
		labels)
}

func (o *factory) HistogramVecVerbose(opt prometheus.HistogramOpts, labels []string) metrics.HistogramVec {
	opt.Namespace, opt.Subsystem = o.namespace, o.subsystem
	return o.registry.HistogramVecVerbose(opt, labels)
}

func (o *factory) SummaryVec(name, help string, labels []string, objectives map[float64]float64) metrics.SummaryVec {
	return o.SummaryVecVerbose(
		prometheus.SummaryOpts{Name: name, Help: help, Objectives: objectives},
		labels)
}

func (o *factory) SummaryVecVerbose(opt prometheus.SummaryOpts, labels []string) metrics.SummaryVec {
	opt.Namespace, opt.Subsystem = o.namespace, o.subsystem
	return o.registry.SummaryVecVerbose(opt, labels)
}

const (
	containerKeyProvider   = "ddnsmetric.factory.provider"
	containerKeyDatasource = "ddnsmetric.factory.datasource"
	containerKeyService    = "ddnsmetric.factory.service"
)

func Pass(ctx context.Context, reg metrics.Registry) {
	container.StoreContext[Factory](ctx, containerKeyProvider,
		NewFactory(reg, Namespace, SubsystemProvider))
	container.StoreContext[Factory](ctx, containerKeyDatasource,
		NewFactory(reg, Namespace, SubsystemDatasource))
	container.StoreContext[Factory](ctx, containerKeyService,
		NewFactory(reg, Namespace, SubsystemService))
}

func FromContext(ctx context.Context, owner adapter.ManagedType) Factory {
	var key string
	switch owner.(type) {
	case adapter.Provider:
		key = containerKeyProvider
	case adapter.Datasource:
		key = containerKeyDatasource
	case adapter.Service:
		key = containerKeyService
	default:
		panic("unknown adapter type")
	}
	f, _ := container.LoadContext[Factory](ctx, key)
	return f
}
