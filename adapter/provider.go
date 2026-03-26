package adapter

import (
	"context"
	"net/netip"
)

type Provider interface {
	Type() string
	Name() string
	Diff(ctx context.Context, domain string, addr []netip.Addr) (bool, error)
	Update(ctx context.Context, domain string, ttl int, addr []netip.Addr) error
}
