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
