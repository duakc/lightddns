package lightddns

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/debug"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

var errDomainNotEnabled = errors.New("not enabled")

type Domain struct {
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
		return nil, errDomainNotEnabled
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

	provider, found := providerManager.Lookup(opt.Provider)
	if !found {
		return nil, &adapter.ManagedNotFoundError{Name: opt.Provider}
	}
	datasource, found := datasourceManager.Lookup(opt.Datasource)
	if !found {
		return nil, &adapter.ManagedNotFoundError{Name: opt.Datasource}
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

func (o *Domain) UpdateOnce(ctx context.Context) error {
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

		logger.Info("no available IP address found")
		return nil
	}

	logger.Debug("found ip", zap.Stringers("ip", netips))

	if err := o.provider.Update(ctx, o.domainName, o.ttl, netips); err != nil {
		return fmt.Errorf("update domain(%s) failed: %w", o.domainName, err)
	}
	return nil
}

func (o *Domain) UpdateLoop(ctx context.Context) {
	ticker := time.NewTicker(o.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := o.UpdateOnce(ctx)
			if err != nil {
				o.logger.Error("update failed", zap.Error(err))
			}
			ticker.Reset(o.updateInterval)
		case <-ctx.Done():
			o.logger.Warn("quited", zap.Error(ctx.Err()))
			return
		}
	}
}
