package lightddns

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/metrics"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"
	"github.com/duakc/mt/debug"
	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/container"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Metrics declared by Domain. Final names are
// "ddns_domain_<leaf>" after ddnsmetric.Factory prefixes them.
const (
	metricActivationTotal       = "activation_total"
	metricUpdateSuccessTotal    = "update_success_total"
	metricUpdateFailureTotal    = "update_failure_total"
	metricNoIPAddressTotal      = "no_ip_address_total"
	metricActualUpdateTotal     = "actual_update_total"
	metricUpdateDurationSeconds = "update_duration_seconds"
	metricLastUpdateTimestamp   = "last_update_timestamp_seconds"
)

type Domain struct {
	taskCtx    context.Context
	taskCancel context.CancelFunc

	closed    chan struct{}
	closeOnce sync.Once
	closeErr  error

	activationCounter    prometheus.Counter
	updateSuccessCounter prometheus.Counter
	updateFailureCounter prometheus.Counter
	noIpAddressCounter   prometheus.Counter
	actualUpdateCounter  prometheus.Counter
	updateDuration       prometheus.Observer
	lastUpdateTimestamp  prometheus.Gauge

	logger         *zap.Logger
	provider       adapter.Provider
	datasource     adapter.Datasource
	domainName     string
	updateInterval time.Duration
	timeout        time.Duration
	ttl            uint32
	ipv4           bool
	ipv6           bool
}

func NewDomain(ctx context.Context, logger *zap.Logger, opt options.DomainOption) (*Domain, error) {
	if !opt.Enabled || len(opt.Domain) == 0 {
		return nil, nil
	}

	opt.IPv4 = opt.IPv4 == opt.IPv6
	opt.IPv6 = opt.IPv4 == opt.IPv6

	datasourceManager := services.Lookup[adapter.DatasourceManager](ctx)
	providerManager := services.Lookup[adapter.ProviderManager](ctx)

	updateInterval := cmp.Or(time.Duration(opt.Interval), constpkg.DefaultDomainUpdateInterval)
	timeout := cmp.Or(time.Duration(opt.Timeout), constpkg.DefaultDomainTimeout)

	switch {
	case timeout > updateInterval:
		return nil, fmt.Errorf("timeout too long (%v > %v)", timeout, updateInterval)
	case updateInterval < time.Second:
		logger.Warn("update interval too short, consider to increase the update interval")
	case timeout < time.Second:
		logger.Warn("timeout too short, consider to increase the timeout")
	}

	provider, providerFound := providerManager.LookupDefault(opt.Provider)
	if !providerFound {
		return nil, &adapter.ProviderNotFoundError{Err: adapter.NewManagedNotFoundError(opt.Provider)}
	}
	datasource, datasourceFound := datasourceManager.LookupDefault(opt.Datasource)
	if !datasourceFound {
		return nil, &adapter.DatasourceNotFoundError{Err: adapter.NewManagedNotFoundError(opt.Datasource)}
	}

	d := &Domain{
		logger:         logger,
		provider:       provider,
		datasource:     datasource,
		domainName:     string(opt.Domain),
		updateInterval: updateInterval,
		timeout:        timeout,
		ttl:            opt.TTL,
		ipv4:           opt.IPv4,
		ipv6:           opt.IPv6,
	}

	d.RegisterMetrics(services.Lookup[metrics.Registry](ctx))

	return d, nil
}

func (o *Domain) Start(ctx context.Context, stage services.Stage) error {
	var err error
	if stage == services.StageStart {
		o.taskCtx, o.taskCancel = context.WithCancel(ctx)
		o.closed = make(chan struct{})
		err = o.Update(o.taskCtx)
		if err != nil {
			return err
		}
		err = o.updateLoop()
		if err != nil {
			return err
		}
	}

	if err = services.Start(ctx, stage, o.provider); err != nil {
		return err
	}

	if err = services.Start(ctx, stage, o.datasource); err != nil {
		return err
	}
	return nil
}

func (o *Domain) Close() error {
	o.closeOnce.Do(func() {
		if o.taskCancel != nil {
			o.taskCancel()
		}
		if o.closed != nil {
			<-o.closed
		}
		o.closeErr = errors.Join(
			o.closeErr,
			services.CloseService(o.provider),
			services.CloseService(o.datasource))
	})
	return o.closeErr
}

func (o *Domain) Update(ctx context.Context) (err error) {
	logger := o.logger
	start := time.Now()
	o.activationCounter.Inc()
	defer func() {
		o.updateDuration.Observe(time.Since(start).Seconds())
		o.lastUpdateTimestamp.Set(float64(time.Now().Unix()))
		if err != nil {
			o.updateFailureCounter.Inc()
		} else {
			o.updateSuccessCounter.Inc()
		}
	}()

	var cont container.Container
	containerProvider := services.Lookup[container.Provider](ctx)
	ctx, cont = containerProvider.New(ctx)
	defer containerProvider.Release(cont)
	cont.IncRef()
	defer cont.DecRef()

	cancelContext, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	netips, err := adapter.MergeDatasources(cancelContext,
		[]adapter.Datasource{o.datasource}, o.ipv4, o.ipv6, true)
	if err != nil {
		return err
	} else if len(netips) == 0 {
		o.noIpAddressCounter.Inc()
		if debug.Enabled {
			return fmt.Errorf("no available IP address found")
		}

		logger.Warn("no available IP address from datasource, skip this update")
		return nil
	}

	logger.Debug("found ip", zap.Stringers("ip", netips))

	changed, err := o.provider.Update(cancelContext, o.domainName, o.ttl, netips)
	if err != nil {
		return fmt.Errorf("update domain(%s) failed: %w", o.domainName, err)
	}
	if changed {
		o.actualUpdateCounter.Inc()
	}
	return nil
}

func (o *Domain) updateLoop() error {
	if o.taskCtx == nil {
		panic("nil task context")
	}
	if mt.Done(o.taskCtx) {
		return o.taskCtx.Err()
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("panic: %v", r)
				}
				o.closeErr = err
			}
			close(o.closed)
		}()
		defer o.taskCancel()
		ctx := o.taskCtx
		logger := o.logger
		ticker := time.NewTicker(o.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				//if causeErr := context.Cause(ctx); causeErr != nil &&
				//	!errors.Is(causeErr, context.Canceled) && !errors.Is(causeErr, signal.NotifyContext()) {
				//	o.closeErr = causeErr
				//}
				//if o.closed != nil {
				//	logger.Warn("quited", zap.Error(o.closeErr))
				//}
				return
			case <-ticker.C:
				err := o.Update(ctx)
				if err != nil {
					logger.Error("update failed", zap.Error(err))
				}
			}
		}
	}()
	return nil
}

func (o *Domain) RegisterMetrics(factory metrics.Factory) {
	labels := []string{constpkg.MetricLabelDomain}
	o.activationCounter = factory.CounterVec(metricActivationTotal,
		"Total number of times a domain update was triggered.", labels).With(o.domainName)
	o.updateSuccessCounter = factory.CounterVec(metricUpdateSuccessTotal,
		"Total number of domain updates that completed without error.", labels).With(o.domainName)
	o.updateFailureCounter = factory.CounterVec(metricUpdateFailureTotal,
		"Total number of domain updates that returned an error.", labels).With(o.domainName)
	o.noIpAddressCounter = factory.CounterVec(metricNoIPAddressTotal,
		"Total number of domain updates skipped because the datasource produced no IP address.",
		labels).With(o.domainName)
	o.actualUpdateCounter = factory.CounterVec(metricActualUpdateTotal,
		"Total number of domain updates that actually mutated DNS records at the provider.",
		labels).With(o.domainName)
	o.updateDuration = factory.HistogramVec(metricUpdateDurationSeconds,
		"Duration of a domain update cycle.", labels, nil).With(o.domainName)
	o.lastUpdateTimestamp = factory.GaugeVec(metricLastUpdateTimestamp,
		"Unix timestamp of the most recent domain update attempt.", labels).With(o.domainName)
}
