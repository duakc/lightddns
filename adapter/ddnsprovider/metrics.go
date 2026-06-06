package ddnsprovider

import (
	"time"

	"github.com/duakc/lightddns/adapter/ddnsmetric"
	"github.com/duakc/lightddns/infra/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type apiMetricsSet struct {
	total    prometheus.Counter
	failure  prometheus.Counter
	duration prometheus.Observer
}

type ApiMetricsRouter struct {
	factory      metrics.Factory
	providerName string

	sets map[string]apiMetricsSet
}

func NewMetricsRouter(factory metrics.Factory, providerName string) *ApiMetricsRouter {
	return &ApiMetricsRouter{
		sets:         make(map[string]apiMetricsSet),
		factory:      factory,
		providerName: providerName,
	}
}

func (r *ApiMetricsRouter) RegisterDefault() {
	var defaults = []string{
		OpDescribeDomains, OpListRecords, OpCreateRecord,
		OpUpdateRecord, OpDeleteRecord,
	}
	r.Register(defaults...)
}

func (r *ApiMetricsRouter) Register(ops ...string) {
	for _, op := range ops {
		r.sets[op] = apiMetricsSet{
			total:    ddnsmetric.ProviderLeaf.RequestTotal(r.factory, r.providerName, op),
			failure:  ddnsmetric.ProviderLeaf.RequestFailure(r.factory, r.providerName, op),
			duration: ddnsmetric.ProviderLeaf.RequestDuration(r.factory, r.providerName, op, nil),
		}
	}
}

func (r *ApiMetricsRouter) RecordAPI(op string) func(errp *error) {
	set, ok := r.sets[op]
	if !ok {
		panic("ddnsmetric: unknown provider op: " + op)
	}
	start := time.Now()
	return func(errp *error) {
		set.total.Inc()
		set.duration.Observe(time.Since(start).Seconds())
		if errp != nil && *errp != nil {
			set.failure.Inc()
		}
	}
}
