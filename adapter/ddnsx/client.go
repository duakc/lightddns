package ddnsx

import (
	"context"

	mDns "github.com/miekg/dns"
)

type DDNSClient[R any] interface {
	DomainSearcher

	Records(ctx context.Context, dnsType mDns.Type) ([]R, error)

	Update(ctx context.Context, records ...R) error
	Delete(ctx context.Context, records ...R) error
	Create(ctx context.Context, records ...R) error
}
