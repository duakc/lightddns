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
	"github.com/duakc/lightddns/infra/lookctx"
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
	datasource []adapter.Datasource

	domainName     string
	updateInterval time.Duration

	ttl  uint32
	ipv4 bool
	ipv6 bool
}

func NewDomain(ctx context.Context, opt options.DomainOption) (*Domain, error) {
	var (
		dataSources    []adapter.Datasource
		provider       adapter.Provider
		updateInterval = constpkg.DefaultUpdateInterval
	)

	if !opt.Enabled || len(opt.Domain) == 0 {
		return nil, errDomainNotEnabled
	}
	if len(opt.DataSource.Value) == 0 {
		return nil, fmt.Errorf("no datasource configured")
	}
	if len(opt.Provider) == 0 {
		return nil, fmt.Errorf("no provider configured")
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
	providerManager := lookctx.Lookup[adapter.ProviderManagerKey, *adapter.ProviderManager](ctx)
	dataSourceManager := lookctx.Lookup[adapter.DatasourceManagerKey, *adapter.DatasourceManager](ctx)

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
		logger:         zaplog.ExtendName(logger, string(opt.Domain)),
		provider:       provider,
		datasource:     dataSources,
		updateInterval: updateInterval,

		domainName: string(opt.Domain),
		ttl:        opt.TTL,
		ipv4:       opt.IPv4,
		ipv6:       opt.IPv6,
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
		zap.String("domain", d.domainName), zap.Stringers("ip", netips))

	if err := d.provider.Update(ctx, d.domainName, d.ttl, netips); err != nil {
		return fmt.Errorf("update domain(%s),provider(%s,%s) failed: %w",
			d.domainName, d.provider.Type(), d.provider.Name(), err)
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
		logger.Sync()
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
		if dualStackDataSource, ok := datasource.(adapter.DatasourceDualStack); ok {
			addresses, err = d.fetchIPDualStack(ctx, dualStackDataSource)
		} else {
			addresses, err = datasource.IP(ctx)
			if err != nil {
				err = newDatasourceError(err, "4/6", datasource)
			}
		}
		if err != nil {
			return nil, err
		}
		netips = append(netips, addresses...)
	}

	netips = common.Filter(netips, func(addr netip.Addr) bool {
		return addr.IsValid() &&
			(d.ipv6 && netxx.IsIPv6(addr) || d.ipv4 && netxx.IsIPv4(addr))
	})
	return common.Distinct(netips), nil
}

func (d *Domain) fetchIPDualStack(ctx context.Context,
	dualStackDataSource adapter.DatasourceDualStack) ([]netip.Addr, error) {
	var netips []netip.Addr
	if d.ipv4 {
		addr, err := dualStackDataSource.IPv4(ctx)
		if err != nil {
			return nil, newDatasourceError(err, "4", dualStackDataSource)
		}
		netips = append(netips, addr...)
	}
	if d.ipv6 {
		addr, err := dualStackDataSource.IPv6(ctx)
		if err != nil {
			return nil, newDatasourceError(err, "6", dualStackDataSource)
		}
		netips = append(netips, addr...)
	}
	return netips, nil
}

type datasourceError struct {
	Err       error
	IPVersion string
	Name      string
	Type      string
}

func newDatasourceError(err error, ipVersion string, ds adapter.Datasource) *datasourceError {
	return &datasourceError{
		Err:       err,
		IPVersion: ipVersion,
		Name:      ds.Name(),
		Type:      ds.Type(),
	}
}

func (e *datasourceError) Error() string {
	return fmt.Sprintf("get ipv%s addresses from datasource(%s,%s) failed: %s",
		e.IPVersion, e.Type, e.Name, e.Err.Error())
}

func (e *datasourceError) Unwrap() error {
	return e.Err
}
