package datasourcex

import (
	"context"
	"errors"
	"net/netip"

	"github.com/duakc/lightddns/adapter"
)

var _ adapter.Datasource = (*LimitedDatasource)(nil)

func NewLimited(raw adapter.Datasource, ipv4, ipv6, fastfail bool) *LimitedDatasource {
	if limitedRaw, isLimited := raw.(*LimitedDatasource); isLimited {
		limitedRaw.IPv4, limitedRaw.IPv6, limitedRaw.FastFail = ipv4, ipv6, fastfail
		return limitedRaw
	}
	return &LimitedDatasource{
		Datasource: raw,
		IPv4:       ipv4,
		IPv6:       ipv6,
		FastFail:   fastfail,
	}
}

type LimitedDatasource struct {
	adapter.Datasource

	IPv4, IPv6 bool
	FastFail   bool
}

func (l *LimitedDatasource) IP(ctx context.Context) ([]netip.Addr, error) {
	if l.IPv4 == l.IPv6 && !l.IPv4 {
		// do nothing
		return []netip.Addr{}, nil
	}

	dualStackDatasource, isDualStack := l.Datasource.(adapter.DatasourceDualStack)

	if !isDualStack {
		list, err := l.Datasource.IP(ctx)
		if err != nil {
			err = newDatasourceError(err, "4/6", l.Datasource)
		}
		return list, err
	}

	var (
		ipv4List, ipv6List []netip.Addr
		err                error
	)
	if l.IPv4 {
		ipv4List, err = dualStackDatasource.IPv4(ctx)
		if err != nil {
			err = newDatasourceError(err, "4", dualStackDatasource)
			if l.FastFail {
				return nil, err
			}
		}
	}

	if l.IPv6 {
		var ipv6Err error
		ipv6List, ipv6Err = dualStackDatasource.IPv6(ctx)
		if ipv6Err != nil {
			err = errors.Join(err, newDatasourceError(ipv6Err, "6", dualStackDatasource))
		}
	}

	return append(ipv4List, ipv6List...), err
}
