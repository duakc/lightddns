package sum

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

const DatasourceType = constpkg.DatasourceGroupTypeSum

func init() {
	adapter.Register(
		adapter.DatasourceRegister,
		DatasourceType,
		New,
	)
}

func New(ctx context.Context, option options.DatasourceGroupSumOption) (adapter.Datasource, error) {
	if len(option.Datasources) == 0 {
		return nil, &adapter.EmptyGroupError{Type: DatasourceType, Name: option.Name}
	}

	logger := option.AbstractDatasourceOption.CreateLogger(
		zaplog.FromContext(ctx))

	datasourceManager := services.Lookup[adapter.DatasourceManager](ctx)
	var datasources []adapter.Datasource

	for i := 0; i < len(option.Datasources); i++ {
		name := option.Datasources[i]
		if datasource, found := datasourceManager.Lookup(name); found {
			datasources = append(datasources, datasource)
		} else {
			return nil, &adapter.ManagedNotFoundError{Name: name}
		}
	}

	sum := &Sum{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),
		logger:              logger,
		datasources:         datasources,
		fastFail:            option.FastFail,
	}
	return sum, nil
}

type Sum struct {
	adapter.AbstractManagedType

	logger      *zap.Logger
	fastFail    bool
	datasources []adapter.Datasource
}

func (s *Sum) IP(ctx context.Context) ([]netip.Addr, error) {
	return s.handle(ctx, true, true)
}

func (s *Sum) IPv4(ctx context.Context) ([]netip.Addr, error) {
	return s.handle(ctx, true, false)
}

func (s *Sum) IPv6(ctx context.Context) ([]netip.Addr, error) {
	return s.handle(ctx, false, true)
}

func (s *Sum) handle(ctx context.Context, ipv4, ipv6 bool) ([]netip.Addr, error) {
	ips, err := adapter.MergeDatasources(ctx, s.datasources, ipv4, ipv6, s.fastFail)
	switch {
	case err != nil:
		return nil, err
	case len(ips) == 0:
		return nil, fmt.Errorf("no IP addresses found")
	}
	return ips, nil
}
