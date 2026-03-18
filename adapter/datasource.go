package adapter

import (
	"context"
	"net/netip"
)

type DataSource interface {
	Type() string
	Name() string
	GetIPv4(context.Context) ([]netip.Addr, error)
	GetIPv6(context.Context) ([]netip.Addr, error)
}
