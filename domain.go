package lightddns

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	datasourcepkg "github.com/duakc/lightddns/datasources"
	"github.com/duakc/lightddns/infra/lookctx"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	providerpkg "github.com/duakc/lightddns/providers"

	"github.com/duakc/mt"

	"go.uber.org/zap"
)

var errDomainNotEnabled = errors.New("not enabled")

type Domain struct {
	logger *zap.Logger

	provider   adapter.Provider
	datasource adapter.Datasource

	domainName     string
	updateInterval time.Duration

	ttl  uint32
	ipv4 bool
	ipv6 bool
}

func NewDomain(ctx context.Context, opt options.DomainOption) (*Domain, error) {
	updateInterval := constpkg.DefaultUpdateInterval

	if !opt.Enabled || len(opt.Domain) == 0 {
		return nil, errDomainNotEnabled
	}
	if len(opt.DataSource) == 0 {
		return nil, fmt.Errorf("missing datasource")
	}
	if len(opt.Provider) == 0 {
		return nil, fmt.Errorf("missing provider")
	}
	if !opt.IPv4 && !opt.IPv6 {
		// enable dual stack default
		opt.IPv4 = true
		opt.IPv6 = true
	}
	if time.Duration(opt.Interval) != 0 {
		updateInterval = time.Duration(opt.Interval)
	}

	logger := lookctx.Lookup[zaplog.LoggerKey, *zap.Logger](ctx)
	dataSourceManager := lookctx.Lookup[adapter.DatasourceManagerKey, *adapter.DatasourceManager](ctx)
	providerManager := lookctx.Lookup[adapter.ProviderManagerKey, *adapter.ProviderManager](ctx)

	var (
		provider   adapter.Provider
		datasource adapter.Datasource
		found      bool
	)

	if provider, found = providerManager.Lookup(opt.Provider); !found {
		return nil, &providerpkg.ProviderNotFoundError{Name: opt.Provider}
	}
	if datasource, found = dataSourceManager.Lookup(opt.DataSource); !found {
		return nil, &datasourcepkg.DatasourceNotFoundError{Name: opt.DataSource}
	}

	return &Domain{
		logger:         logger.Named(string(opt.Domain)),
		provider:       provider,
		datasource:     datasource,
		updateInterval: updateInterval,

		domainName: string(opt.Domain),
		ttl:        opt.TTL,
		ipv4:       opt.IPv4,
		ipv6:       opt.IPv6,
	}, nil
}

func (o *Domain) UpdateOnce(ctx context.Context) error {
	// TODO: here needs some optimization
	logger := o.logger
	var cancel context.CancelFunc
	ctx, cancel = mt.Timeout(ctx, o.updateInterval)
	defer cancel()

	netips, err := adapter.MergeDatasources(ctx,
		[]adapter.Datasource{o.datasource}, o.ipv4, o.ipv6, true)
	if err != nil {
		return err
	}
	logger.Debug("found ip",
		zap.String("domain", o.domainName), zap.Stringers("ip", netips))

	if err := o.provider.Update(ctx, o.domainName, o.ttl, netips); err != nil {
		return fmt.Errorf("update domain(%s) failed: %w",
			o.domainName, err)
	}
	return nil
}

func (o *Domain) UpdateLoop(ctx context.Context) {
	// TODO: here needs some optimization
	logger := o.logger
	ticker := time.NewTicker(o.updateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			deadlineCtx, cancel := context.WithTimeout(ctx, o.updateInterval)
			err := o.UpdateOnce(deadlineCtx)
			if err != nil {
				logger.Error("update failed", zap.Error(err))
			}
			cancel()
		case <-ctx.Done():
			logger.Warn("quited", zap.Error(ctx.Err()))
			return
		}
		ticker.Reset(o.updateInterval)
	}
}
