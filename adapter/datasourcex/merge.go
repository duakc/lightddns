package datasourcex

import (
	"context"
	"errors"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
)

func MergeDualStackDatasourceIP(ctx context.Context, s adapter.DatasourceDualStack) ([]netip.Addr, error) {
	return NewLimited(s, true, true, true).IP(ctx)
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
