package failover

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"sync"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

var _ adapter.DatasourceGroup = (*FailOver)(nil)

const DatasourceType = constpkg.DatasourceGroupTypeFailover

func init() {
	adapter.Register(
		adapter.DatasourceRegister,
		DatasourceType,
		New,
	)
}

func New(ctx context.Context, logger *zap.Logger, option options.DatasourceGroupFailoverOption) (adapter.Datasource, error) {
	if len(option.Datasources) == 0 {
		return nil, &adapter.EmptyGroupError{Type: DatasourceType, Name: option.Name}
	}

	failover := &FailOver{
		AbstractManagedType:    adapter.NewManagedType(DatasourceType, option.Name),
		DatasourceGroupBuilder: adapter.NewDatasourceGroupBuild(option.Datasources),

		logger: logger,
	}

	return failover, nil
}

type FailOver struct {
	adapter.AbstractManagedType
	adapter.DatasourceGroupBuilder

	logger      *zap.Logger
	datasources []adapter.Datasource

	lastSuccess int
	access      sync.Mutex
}

func (f *FailOver) WithManager(manager adapter.DatasourceManager) error {
	var err error
	f.datasources, err = f.DatasourceGroupBuilder.Build(manager)
	return err
}

func (f *FailOver) Start(ctx context.Context, stage services.Stage) error {
	for i := 0; i < len(f.datasources); i++ {
		err := services.Start(ctx, stage, f.datasources[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FailOver) Close() error {
	var err error

	for i := 0; i < len(f.datasources); i++ {
		err = errors.Join(err, services.CloseService(f.datasources))
	}
	return err
}

func (f *FailOver) IP(ctx context.Context) ([]netip.Addr, error) {
	return f.handle(ctx, true, true)
}

func (f *FailOver) IPv4(ctx context.Context) ([]netip.Addr, error) {
	return f.handle(ctx, true, false)
}

func (f *FailOver) IPv6(ctx context.Context) ([]netip.Addr, error) {
	return f.handle(ctx, false, true)
}

func (f *FailOver) handle(ctx context.Context, ipv4, ipv6 bool) ([]netip.Addr, error) {
	const onceFailPercentage = 2

	if !ipv4 && !ipv6 {
		panic("unexcepted")
	}
	logger := f.logger
	f.access.Lock()
	defer f.access.Unlock()

	for walked := 0; walked <= len(f.datasources)/onceFailPercentage; walked++ {
		var err error
		var ips []netip.Addr
		if f.lastSuccess >= len(f.datasources) {
			f.lastSuccess = 0
		}
		succeedDatasource := f.datasources[f.lastSuccess]
		if ipv6 && ipv4 {
			ips, err = succeedDatasource.IP(ctx)
		} else if dualStack, isDualStack := succeedDatasource.(adapter.DatasourceDualStack); isDualStack {
			switch {
			case ipv4:
				ips, err = dualStack.IPv4(ctx)
			case ipv6:
				ips, err = dualStack.IPv6(ctx)
			}
		} else {
			ips, err = succeedDatasource.IP(ctx)
			ips = netool.FilterAddress(ips, ipv4, ipv6)
		}
		if err == nil {
			f.lastSuccess = (f.lastSuccess + walked) % len(f.datasources)
			return ips, nil
		}
		logger.Warn("failover",
			zap.Error(err),
			zap.String("failed_datasource", succeedDatasource.Name()),
			zap.String("failed_datasource_type", succeedDatasource.Type()),
			zap.Int("walked", walked))
		f.lastSuccess++
		if mt.Done(ctx) {
			if walked+1 < len(f.datasources)/onceFailPercentage {
				logger.Info("failover retry continued, but context deadline exceeded")
			}
			return nil, ctx.Err()
		}
	}
	return nil, fmt.Errorf("failover group all datasources failed")
}
