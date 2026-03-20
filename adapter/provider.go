package adapter

import (
	"context"
	"net/netip"
)

type Provider interface {
	Type() string
	Name() string
	Diff(ctx context.Context, addr []netip.Addr) (bool, error)
	Update(ctx context.Context, addr []netip.Addr) error
}
