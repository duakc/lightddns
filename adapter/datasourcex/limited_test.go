package datasourcex

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/duakc/lightddns/adapter"

	"github.com/stretchr/testify/require"
)

func TestLimitedDatasourceDualStackBothFamiliesDoesNotCallIP(t *testing.T) {
	ds := &dualStackDatasource{
		AbstractManagedType: adapter.NewManagedType("test", "dual"),
		ipv4:                []netip.Addr{netip.MustParseAddr("192.0.2.1")},
		ipv6:                []netip.Addr{netip.MustParseAddr("2001:db8::1")},
	}

	got, err := NewLimited(ds, true, true, true).IP(context.Background())
	require.NoError(t, err)
	require.Equal(t, []netip.Addr{
		netip.MustParseAddr("192.0.2.1"),
		netip.MustParseAddr("2001:db8::1"),
	}, got)
	require.False(t, ds.ipCalled)
}

type dualStackDatasource struct {
	adapter.AbstractManagedType

	ipv4 []netip.Addr
	ipv6 []netip.Addr

	ipCalled bool
}

func (d *dualStackDatasource) IP(ctx context.Context) ([]netip.Addr, error) {
	d.ipCalled = true
	return nil, errors.New("IP should not be called for dual-stack limited merge")
}

func (d *dualStackDatasource) IPv4(ctx context.Context) ([]netip.Addr, error) {
	return d.ipv4, nil
}

func (d *dualStackDatasource) IPv6(ctx context.Context) ([]netip.Addr, error) {
	return d.ipv6, nil
}
