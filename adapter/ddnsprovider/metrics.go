package ddnsprovider

import (
	"time"

	"github.com/duakc/lightddns/adapter/ddnsmetric"

	"github.com/prometheus/client_golang/prometheus"
)

type apiMetricsSet struct {
	total    prometheus.Counter
	failure  prometheus.Counter
	duration prometheus.Observer
}

type ApiMetricsRouter struct {
	factory      ddnsmetric.ProviderFactory
	providerName string

	sets map[string]apiMetricsSet
}

func NewMetricsRouter(factory ddnsmetric.ProviderFactory, providerName string) *ApiMetricsRouter {
	return &ApiMetricsRouter{
		sets:         make(map[string]apiMetricsSet),
		factory:      factory,
		providerName: providerName,
	}
}

func (r *ApiMetricsRouter) RegisterDefault() {
	defaults := []string{
		OpDescribeDomains, OpListRecords, OpCreateRecord,
		OpUpdateRecord, OpDeleteRecord,
	}
	r.Register(defaults...)
}

func (r *ApiMetricsRouter) Register(ops ...string) {
	for _, op := range ops {
		r.sets[op] = apiMetricsSet{
			total:    r.factory.RequestTotal(r.providerName, op),
			failure:  r.factory.RequestFailure(r.providerName, op),
			duration: r.factory.RequestDuration(r.providerName, op, nil),
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
