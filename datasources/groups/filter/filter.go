package filter

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/datasourcex"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/services"

	"go.uber.org/zap"
	"go4.org/netipx"
)

const DatasourceType = constpkg.DatasourceGroupFilter

func init() {
	adapter.Register(
		adapter.DatasourceRegister,
		DatasourceType,
		New,
	)
}

var _ adapter.DatasourceDualStack = (*Filter)(nil)

type Filter struct {
	adapter.AbstractManagedType

	logger *zap.Logger

	datasources []adapter.Datasource

	prefixes *netipx.IPSet

	invert bool
}

func New(ctx context.Context, logger *zap.Logger, option options.DatasourceGroupFilterOption) (adapter.Datasource, error) {
	if len(option.Prefixes.Value) == 0 {
		return nil, fmt.Errorf("empty prefixes")
	}
	var (
		datasources []adapter.Datasource
		err         error
	)

	datasources, err = datasourcex.Lookup(services.Lookup[adapter.DatasourceManager](ctx), option.Datasources...)
	if err != nil {
		return nil, err
	}

	// build prefixes
	var (
		ipSetBuilder netipx.IPSetBuilder
		ipSet        *netipx.IPSet
	)
	for _, prefix := range option.Prefixes.Value {
		ipSetBuilder.AddPrefix(netip.Prefix(prefix))
	}

	ipSet, err = ipSetBuilder.IPSet()
	if err != nil {
		return nil, fmt.Errorf("build ip set: %w", err)
	}

	return &Filter{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),
		logger:              logger,
		datasources:         datasources,
		prefixes:            ipSet,
		invert:              option.Invert,
	}, nil
}

func (f *Filter) IP(ctx context.Context) ([]netip.Addr, error) {
	collectedIP, err := datasourcex.MergeDatasources(ctx, f.datasources, true, true, true)
	if err != nil {
		return nil, err
	}
	return f.filterOut(collectedIP), nil
}

func (f *Filter) IPv4(ctx context.Context) ([]netip.Addr, error) {
	collectedIP, err := datasourcex.MergeDatasources(ctx, f.datasources, true, false, true)
	if err != nil {
		return nil, err
	}
	return f.filterOut(collectedIP), nil
}

func (f *Filter) IPv6(ctx context.Context) ([]netip.Addr, error) {
	collectedIP, err := datasourcex.MergeDatasources(ctx, f.datasources, false, true, true)
	if err != nil {
		return nil, err
	}
	return f.filterOut(collectedIP), nil
}

func (f *Filter) filterOut(ips []netip.Addr) []netip.Addr {
	var res []netip.Addr
	logger := f.logger
	for _, ip := range ips {
		logger.Debug("filter",
			zap.Stringer("filter_ip", ip),
		)
		if f.prefixes.Contains(ip) != f.invert {
			logger.Info("filter in",
				zap.Stringer("filter_ip", ip),
			)
			res = append(res, ip)
		}
	}

	return res
}
