package failover

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	"github.com/duakc/lightddns/adapter"
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/netool"
	"github.com/duakc/lightddns/infra/zaplog"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"
	"github.com/duakc/mt/services"

	"go.uber.org/zap"
)

const DatasourceType = constpkg.DatasourceGroupTypeFailover

func init() {
	adapter.Register(
		adapter.DatasourceRegister,
		DatasourceType,
		New,
	)
}

func New(ctx context.Context, option options.DatasourceGroupFailoverOption) (adapter.Datasource, error) {
	if len(option.Datasources) == 0 {
		return nil, &adapter.EmptyGroupError{Type: DatasourceType, Name: option.Name}
	}
	logger := option.AbstractDatasourceOption.CreateLogger(
		zaplog.FromContext(ctx))

	var datasources []adapter.Datasource
	manager := services.Lookup[adapter.DatasourceManager](ctx)
	for i := 0; i < len(option.Datasources); i++ {
		name := option.Datasources[i]
		if baseDatasource, found := manager.Lookup(name); found {
			datasources = append(datasources, baseDatasource)
		} else {
			return nil, &adapter.ManagedNotFoundError{Name: name}
		}
	}

	failover := &FailOver{
		AbstractManagedType: adapter.NewManagedType(DatasourceType, option.Name),
		logger:              logger,
		datasources:         datasources,
		lastSuccess:         0,
	}
	return failover, nil
}

type FailOver struct {
	adapter.AbstractManagedType

	logger      *zap.Logger
	datasources []adapter.Datasource
	lastSuccess int
	access      sync.Mutex
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
