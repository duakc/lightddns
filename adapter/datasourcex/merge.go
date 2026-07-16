package datasourcex

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
)

func MergeDualStackDatasourceIP(ctx context.Context, s adapter.DatasourceDualStack) ([]netip.Addr, error) {
	ipv4, err := s.IPv4(ctx)
	if err != nil {
		return nil, fmt.Errorf("ipv4: %w", err)
	}

	ipv6, err := s.IPv6(ctx)
	if err != nil {
		return nil, fmt.Errorf("ipv6: %w", err)
	}
	return append(ipv4, ipv6...), nil
}

func MergeDatasources(ctx context.Context, datasources []adapter.Datasource, ipv4, ipv6, fastfail bool) ([]netip.Addr, error) {
	var (
		iplist []netip.Addr
		err    error
	)
	for _, ds := range datasources {
		limitedDAtasource := NewLimited(ds, ipv4, ipv6, fastfail)
		resultIp, resultErr := limitedDAtasource.IP(ctx)
		if resultErr != nil {
			err = errors.Join(err, resultErr)
		}
		if fastfail && err != nil {
			return nil, err
		}
		iplist = append(iplist, resultIp...)
	}

	return iplist, err
}
