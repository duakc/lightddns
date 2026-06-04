package sum

import (
	"context"
	"errors"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
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

func New(ctx context.Context, logger *zap.Logger, option options.DatasourceGroupSumOption) (adapter.Datasource, error) {
	if len(option.Datasources) == 0 {
		return nil, &adapter.EmptyGroupError{Type: DatasourceType, Name: option.Name}
	}

	sum := &Sum{
		AbstractManagedType:    adapter.NewManagedType(DatasourceType, option.Name),
		DatasourceGroupBuilder: adapter.NewDatasourceGroupBuild(option.Datasources),

		fastFail: option.FastFail,
		logger:   logger,
	}

	return sum, nil
}

var _ adapter.DatasourceGroup = (*Sum)(nil)

type Sum struct {
	adapter.AbstractManagedType
	adapter.DatasourceGroupBuilder

	logger      *zap.Logger
	fastFail    bool
	datasources []adapter.Datasource
}

func (s *Sum) WithManager(manager adapter.DatasourceManager) error {
	var err error
	s.datasources, err = s.DatasourceGroupBuilder.Build(manager)
	return err
}

func (s *Sum) Start(ctx context.Context, stage services.Stage) error {
	for i := 0; i < len(s.datasources); i++ {
		err := services.Start(ctx, stage, s.datasources[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sum) Close() error {
	var err error
	for i := 0; i < len(s.datasources); i++ {
		err = errors.Join(err, services.CloseService(s.datasources[i]))
	}
	return err
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
	return adapter.MergeDatasources(ctx, s.datasources, ipv4, ipv6, s.fastFail)
}
