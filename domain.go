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
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"
	"github.com/duakc/mt/debug"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

type Domain struct {
	taskCtx    context.Context
	taskCancel context.CancelFunc

	closed    chan struct{}
	closeOnce sync.Once
	closeErr  error

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

func NewDomain(ctx context.Context, opt options.DomainOption) (*Domain, error) {
	if !opt.Enabled || len(opt.Domain) == 0 {
		return nil, nil
	}
	if len(opt.Datasource) == 0 {
		return nil, fmt.Errorf("missing datasource")
	}
	if len(opt.Provider) == 0 {
		return nil, fmt.Errorf("missing provider")
	}
	if !opt.IPv4 && !opt.IPv6 {
		opt.IPv4 = true
		opt.IPv6 = true
	}

	logger := services.LookupPtr[zap.Logger](ctx).Named(string(opt.Domain))
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

	provider, providerFound := providerManager.Lookup(opt.Provider)
	if !providerFound {
		return nil, &adapter.ProviderNotFoundError{ManagedNotFoundError: adapter.NewManagedNotFoundError(opt.Provider)}
	}
	datasource, datasourceFound := datasourceManager.Lookup(opt.Datasource)
	if !datasourceFound {
		return nil, &adapter.DatasourceNotFoundError{ManagedNotFoundError: adapter.NewManagedNotFoundError(opt.Datasource)}
	}

	return &Domain{
		logger:         logger,
		provider:       provider,
		datasource:     datasource,
		domainName:     string(opt.Domain),
		updateInterval: updateInterval,
		timeout:        timeout,
		ttl:            opt.TTL,
		ipv4:           opt.IPv4,
		ipv6:           opt.IPv6,
	}, nil
}

func (o *Domain) Start(ctx context.Context, stage services.Stage) error {
	switch stage {
	case services.StagePreStart, services.StagePostStart:
		if err := services.Start(ctx, stage, o.provider); err != nil {
			return err
		}
		return services.Start(ctx, stage, o.datasource)
	case services.StageStart:
		o.taskCtx, o.taskCancel = context.WithCancel(ctx)
		o.closed = make(chan struct{})
		err := o.update(o.taskCtx)
		if err != nil {
			return err
		}
		return o.updateLoop()
	default:
		panic("unknown stage: " + stage.String())
	}
}

func (o *Domain) Close() error {
	o.closeOnce.Do(func() {
		o.taskCancel()
		<-o.closed
		o.closeErr = errors.Join(
			o.closeErr,
			services.CloseService(o.provider),
			services.CloseService(o.datasource))
	})
	return o.closeErr
}

func (o *Domain) update(ctx context.Context) error {
	logger := o.logger

	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	netips, err := adapter.MergeDatasources(ctx,
		[]adapter.Datasource{o.datasource}, o.ipv4, o.ipv6, true)
	if err != nil {
		return err
	} else if len(netips) == 0 {
		if debug.Enabled {
			return fmt.Errorf("no available IP address found")
		}

		logger.Info("no available IP address from datasource, skip this update")
		return nil
	}

	logger.Debug("found ip", zap.Stringers("ip", netips))

	if err := o.provider.Update(ctx, o.domainName, o.ttl, netips); err != nil {
		return fmt.Errorf("update domain(%s) failed: %w", o.domainName, err)
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
				logger.Warn("quited", zap.Error(ctx.Err()))
				if causeErr := context.Cause(ctx); causeErr != nil && !errors.Is(causeErr, context.Canceled) {
					o.closeErr = causeErr
				}
				return
			case <-ticker.C:
				err := o.update(ctx)
				if err != nil {
					logger.Error("update failed", zap.Error(err))
				}
			}
		}
	}()
	return nil
}
