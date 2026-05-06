package sum

import (
	"context"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	datasourcepkg "github.com/duakc/lightddns/datasources"
	"github.com/duakc/lightddns/infra/lookctx"
	"github.com/duakc/lightddns/options"

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
		return nil, &datasourcepkg.EmptyGroupError{Type: DatasourceType, Name: option.Name}
	}

	logger := datasourcepkg.NewLogger(
		lookctx.LookupPtr[zap.Logger](ctx),
		option.AbstractDatasourceOption)

	datasourceManager := lookctx.Lookup[adapter.DatasourceManager](ctx)
	var datasources []adapter.Datasource

	for i := 0; i < len(option.Datasources); i++ {
		name := option.Datasources[i]
		if datasource, found := datasourceManager.Lookup(name); found {
			datasources = append(datasources, datasource)
		} else {
			return nil, &datasourcepkg.DatasourceNotFoundError{Name: name}
		}
	}

	sum := &Sum{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),
		logger:              logger,
		datasources:         datasources,
		fastFail:            !option.IgnoreDownstreamError,
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
	// TODO: here needs some optimization

	logger := s.logger
	ips, err := adapter.MergeDatasources(ctx, s.datasources, ipv4, ipv6, s.fastFail)
	if err != nil && s.fastFail {
		return nil, err
	}
	if err != nil {
		logger.Warn("an error occurred with DatasourceGroupSumOption.IgnoreDownstreamError enabled",
			zap.Error(err), zap.Int("len(ips)", len(ips)))
		// even if the DatasourceGroupSumOption.IgnoreDownstreamError is enabled ,
		// but with empty ip list returned may case this domain unaccessible.
		// so we must return an error here if len(ips) == 0.
		if len(ips) != 0 {
			err = nil
		}
	}
	return ips, err
}
