package adapter

import (
	"context"
	"net/netip"
)

type Datasource interface {
	ManagedType
	IP(context.Context) ([]netip.Addr, error)
}

type DatasourceDualStack interface {
	Datasource
	IPv4(ctx context.Context) ([]netip.Addr, error)
	IPv6(ctx context.Context) ([]netip.Addr, error)
}

type (
	DatasourceManager = Manager[Datasource]
)

var DatasourceRegister = NewRegister[Datasource]()
