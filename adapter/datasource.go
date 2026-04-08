package adapter

import (
	"context"
	"net/netip"
)

type Datasource interface {
	managedType
	IP(context.Context) ([]netip.Addr, error)
}

type DatasourceDualStack interface {
	Datasource
	IPv4(ctx context.Context) ([]netip.Addr, error)
	IPv6(ctx context.Context) ([]netip.Addr, error)
}

type DatasourceManager = DefaultManager[Datasource]
type DatasourceManagerKey struct{}

var DatasourceRegister = NewRegister[Datasource]()

func MergeDualStackDatasourceIP(ctx context.Context, s DatasourceDualStack) ([]netip.Addr, error) {
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
