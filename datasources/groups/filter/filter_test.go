package filter

import (
	"context"
	"net/netip"
	"testing"

	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/services"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewInitializesRuleMatchCache(t *testing.T) {
	ctx := services.NewRegistry(context.Background(), services.NewDefaultRegistry())
	manager := adapter.NewManager[adapter.Datasource](adapter.DatasourceRegister)
	ctx = services.Store[adapter.DatasourceManager](ctx, manager)

	datasource, err := New(ctx, zap.NewNop(), options.DatasourceGroupFilterOption{
		Rules: []options.DatasourceGroupFilterRuleOption{
			{
				Prefixes: []badyaml.Prefix{
					badyaml.Prefix(netip.MustParsePrefix("192.0.2.0/24")),
				},
			},
		},
	})
	require.NoError(t, err)

	filter := datasource.(*Filter)
	require.NotNil(t, filter.ruleMatchCache)

	filter.datasources = []adapter.Datasource{
		&testDatasource{
			addresses: []netip.Addr{
				netip.MustParseAddr("192.0.2.1"),
				netip.MustParseAddr("198.51.100.1"),
			},
		},
	}

	want := []netip.Addr{netip.MustParseAddr("192.0.2.1")}
	got, err := filter.IPv4(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)

	// The second call exercises the cache hit path as well as the first-call
	// initialization path that previously dereferenced a nil cache.
	got, err = filter.IPv4(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

type testDatasource struct {
	adapter.AbstractManagedType
	addresses []netip.Addr
}

func (d *testDatasource) IP(context.Context) ([]netip.Addr, error) {
	return d.addresses, nil
}
