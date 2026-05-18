package lightddns

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/lookctx"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

var errDomainNotEnabled = errors.New("not enabled")

type Domain struct {
	logger         *zap.Logger
	provider       adapter.Provider
	datasource     adapter.Datasource
	domainName     string
	updateInterval time.Duration
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

	updateInterval := constpkg.DefaultUpdateInterval
	if opt.Interval > 0 {
		updateInterval = time.Duration(opt.Interval)
	}

	logger := lookctx.LookupPtr[zap.Logger](ctx).Named(string(opt.Domain))
	datasourceManager := lookctx.Lookup[adapter.DatasourceManager](ctx)
	providerManager := lookctx.Lookup[adapter.ProviderManager](ctx)

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
		ttl:            opt.TTL,
		ipv4:           opt.IPv4,
		ipv6:           opt.IPv6,
	}, nil
}

func (o *Domain) UpdateOnce(ctx context.Context) error {
	ctx, cancel := mt.Timeout(ctx, o.updateInterval)
	defer cancel()

	netips, err := adapter.MergeDatasources(ctx,
		[]adapter.Datasource{o.datasource}, o.ipv4, o.ipv6, true)
	if err != nil {
		return err
	}
	o.logger.Debug("found ip",
		zap.String("domain", o.domainName),
		zap.Stringers("ip", netips))

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
			if err := o.UpdateOnce(ctx); err != nil {
				o.logger.Error("update failed", zap.Error(err))
			}
		case <-ctx.Done():
			o.logger.Warn("quited", zap.Error(ctx.Err()))
			return
		}
	}
}
