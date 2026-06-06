package ddnsmetric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// providerAPISet holds the prometheus collectors bound to one (provider name, op)
// pair. Resolved once at registration so RecordAPI doesn't pay vec.With's hash
// per call — the DDNS poll loop hits these on every request.
type providerAPISet struct {
	total    prometheus.Counter
	failure  prometheus.Counter
	duration prometheus.Observer
}

// ProviderAPIRouter dispatches per-op API metric updates for one provider
// client instance. The op set is fixed at construction; RecordAPI panics on an
// unknown op so typos surface at the first call site exercised after startup.
type ProviderAPIRouter struct {
	sets map[string]providerAPISet
}

// NewRouter builds the router for a provider client. Ops listed here become
// the only legal labels for RecordAPI.
func (providerLeaf) NewRouter(factory Factory, name string, ops []string) *ProviderAPIRouter {
	sets := make(map[string]providerAPISet, len(ops))
	for _, op := range ops {
		sets[op] = providerAPISet{
			total:    ProviderLeaf.RequestTotal(factory, name, op),
			failure:  ProviderLeaf.RequestFailure(factory, name, op),
			duration: ProviderLeaf.RequestDuration(factory, name, op, nil),
		}
	}
	return &ProviderAPIRouter{sets: sets}
}

func (r *ProviderAPIRouter) RecordAPI(op string) func(errp *error) {
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
