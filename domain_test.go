package lightddns

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/adapter/metricx"
	"github.com/duakc/lightddns/infra/metrics"

	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/container"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDomainStartOnceUpdatesWithoutStartingLoop(t *testing.T) {
	provider := &domainTestProvider{}
	domain := &Domain{
		logger:         zap.NewNop(),
		provider:       provider,
		datasource:     &domainTestDatasource{},
		domainName:     "example.com",
		updateInterval: time.Hour,
		timeout:        time.Second,
		ipv4:           true,
		once:           true,
	}
	domain.RegisterMetrics(metrics.NewNameFactory(
		metrics.New(false), metricx.Namespace, metricx.SubsystemDomain,
	))

	ctx := services.Store[container.Provider](
		context.Background(), container.NewDefaultProvider())
	require.NoError(t, domain.Start(ctx, services.StageStart))
	require.Equal(t, 1, provider.updateCount)
	require.False(t, domain.loopStarted)
	require.Nil(t, domain.taskCancel)
	require.Nil(t, domain.closed)

	// Close must not wait for a loop that was intentionally never started.
	require.NoError(t, domain.Close())
}

type domainTestDatasource struct {
	adapter.AbstractManagedType
}

func (d *domainTestDatasource) IP(context.Context) ([]netip.Addr, error) {
	return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil
}

type domainTestProvider struct {
	adapter.AbstractManagedType
	updateCount int
}

func (p *domainTestProvider) Diff(context.Context, string, uint32, []netip.Addr) (bool, error) {
	return false, nil
}

func (p *domainTestProvider) Update(context.Context, string, uint32, []netip.Addr) (bool, error) {
	p.updateCount++
	return true, nil
}

func (p *domainTestProvider) Close() error { return nil }
