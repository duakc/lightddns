package adapter

import (
	"context"
	"net/netip"
)

type DataSource interface {
	managedType
	IP(context.Context) ([]netip.Addr, error)
}

type DataSourceDualStack interface {
	DataSource
	IPv4(ctx context.Context) ([]netip.Addr, error)
	IPv6(ctx context.Context) ([]netip.Addr, error)
}

type DataSourceManager = DefaultManager[DataSource]
type DataSourceManagerKey struct{}

var DataSourceRegister = NewRegister[DataSource]()

func MergeDualStackDatasourceIP(ctx context.Context, s DataSourceDualStack) ([]netip.Addr, error) {
	ipv4, err := s.IPv4(ctx)
	if err != nil {
		return nil, err
	}

	ipv6, err := s.IPv6(ctx)
	if err != nil {
		return nil, err
	}
	return append(ipv4, ipv6...), nil
}
