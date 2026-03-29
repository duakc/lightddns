package lightddns

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/common"
	"github.com/duakc/lightddns/infra/ctxservice"
	"github.com/duakc/lightddns/infra/netxx"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"
	"go.uber.org/zap"
)

var (
	errDomainNotEnabled = errors.New("not enabled")
)

type Domain struct {
	logger *zap.Logger

	provider   adapter.Provider
	datasource []adapter.DataSource

	domainNames string
	ttl         int
	ipv4        bool
	ipv6        bool

	updateInterval time.Duration
}

func NewDomain(ctx context.Context, opt options.OptionDomain) (*Domain, error) {
	if !opt.Enabled || len(opt.Domain) == 0 {
		return nil, errDomainNotEnabled
	}
	if len(opt.DataSource.Value) == 0 {
		return nil, fmt.Errorf("no datasource configured")
	}
	if len(opt.Provider) == 0 {
		return nil, fmt.Errorf("no provider configured")
	}
	if opt.TTL < 0 {
		opt.TTL = 0
	}
	if !opt.IPv4 && !opt.IPv6 {
		// enable dual stack default
		opt.IPv4 = true
		opt.IPv6 = true
	}
	if !netxx.IsDomainName(opt.Domain) {
		return nil, fmt.Errorf("domain %s not a valid domain name", opt.Domain)
	}
	var (
		dataSources    []adapter.DataSource
		provider       adapter.Provider
		updateInterval time.Duration = constpkg.DefaultUpdateInterval
	)

	logger := ctxservice.Lookup[*zap.Logger](ctx, common.Zero[zaplog.LoggerKey]())
	providerManager := ctxservice.Lookup[*adapter.ProviderManager](ctx, common.Zero[adapter.ProviderManagerKey]())
	dataSourceManager := ctxservice.Lookup[*adapter.DataSourceManager](ctx, common.Zero[adapter.DataSourceManagerKey]())

	if opt.Interval != "" {
		updateIntervalParsed, err := time.ParseDuration(opt.Interval)
		if err != nil {
			return nil, fmt.Errorf("ParseDuration: %w", err)
		}
		if updateIntervalParsed < 5*time.Second {
			return nil, fmt.Errorf("update interval can not lower than 5s")
		}
		updateInterval = updateIntervalParsed
	}

	for i := 0; i < len(opt.DataSource.Value); i++ {
		dataSourceName := opt.DataSource.Value[i]
		if dataSource, ok := dataSourceManager.Lookup(dataSourceName); !ok {
			return nil, fmt.Errorf("datasource %s not found", dataSourceName)
		} else {
			dataSources = append(dataSources, dataSource)
		}
	}

	if providerLookup, ok := providerManager.Lookup(opt.Provider); !ok {
		return nil, fmt.Errorf("provider %s not found", opt.Provider)
	} else {
		provider = providerLookup
	}

	return &Domain{
		logger:         logger,
		provider:       provider,
		datasource:     dataSources,
		updateInterval: updateInterval,

		domainNames: opt.Domain,
		ttl:         opt.TTL,
		ipv4:        opt.IPv6,
		ipv6:        opt.IPv6,
	}, nil
}

func (d *Domain) UpdateOnce(ctx context.Context) error {
	logger := d.logger
	defer logger.Sync()

	var cancel context.CancelFunc
	ctx, cancel = common.ContextWithTimeout(ctx, d.updateInterval)
	defer cancel()

	netips, err := d.fetchIP(ctx)
	if err != nil {
		return fmt.Errorf("fetchIP: %w", err)
	}
	logger.Debug("found ip",
		zap.String("domain", d.domainNames), zap.Stringers("ip", netips))

	if err := d.provider.Update(ctx, d.domainNames, d.ttl, netips); err != nil {
		return fmt.Errorf("update %s: %w", d.domainNames, err)
	}
	return nil
}

func (d *Domain) UpdateLoop(ctx context.Context) {
	logger := d.logger
	defer logger.Sync()
	ticker := time.NewTicker(d.updateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			deadlineCtx, cancel := context.WithTimeout(ctx, d.updateInterval)
			err := d.UpdateOnce(deadlineCtx)
			if err != nil {
				logger.Error("update failed", zap.Error(err))
			}
			cancel()
		case <-ctx.Done():
			logger.Warn("quited", zap.Error(ctx.Err()))
			return
		}
		ticker.Reset(d.updateInterval)
	}
}

func (d *Domain) fetchIP(ctx context.Context) ([]netip.Addr, error) {
	var netips []netip.Addr

	for i := 0; i < len(d.datasource); i++ {

		var (
			datasource = d.datasource[i]
			addresses  []netip.Addr
			err        error
		)
		if dualStackDataSource, ok := datasource.(adapter.DataSourceDualStack); ok {
			addresses, err = d.fetchIPDualStack(ctx, dualStackDataSource)
		} else {
			addresses, err = datasource.IP(ctx)
			if err != nil {
				err = &datasourceError{
					Err:  err,
					IPv4: true,
					IPv6: true,
					Name: datasource.Name(),
				}
			}
		}
		if err != nil {
			return nil, err
		}
		netips = append(netips, addresses...)
	}
	return netips, nil
}

func (d *Domain) fetchIPDualStack(ctx context.Context,
	dualStackDataSource adapter.DataSourceDualStack) ([]netip.Addr, error) {
	var netips []netip.Addr
	if d.ipv4 {
		addr, err := dualStackDataSource.IPv4(ctx)
		if err != nil {
			return nil, &datasourceError{
				Err:  err,
				IPv4: true,
				Name: dualStackDataSource.Name(),
			}
		}
		netips = append(netips, addr...)
	}
	if d.ipv6 {
		addr, err := dualStackDataSource.IPv4(ctx)
		if err != nil {
			return nil, &datasourceError{
				Err:  err,
				IPv6: true,
				Name: dualStackDataSource.Name(),
			}
		}
		netips = append(netips, addr...)
	}
	return netips, nil
}

type datasourceError struct {
	Err  error
	IPv4 bool
	IPv6 bool
	Name string
}

func (e *datasourceError) Error() string {
	ipVersion := ""
	switch {
	case e.IPv4 && e.IPv6:
		// nop
	case e.IPv6:
		ipVersion = "6"
	case e.IPv4:
		ipVersion = "4"
	}
	return fmt.Sprintf("get ipv%s addresses from datasource(%s) failed: %s",
		ipVersion, e.Name, e.Err.Error())
}

func (e *datasourceError) Unwrap() error {
	return e.Err
}
