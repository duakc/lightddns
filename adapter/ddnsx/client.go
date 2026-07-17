package ddnsx

import (
	"context"
	"net/netip"
)

type RecordType string

const (
	RecordTypeA    RecordType = "A"
	RecordTypeAAAA RecordType = "AAAA"
)

func (t RecordType) String() string {
	return string(t)
}

type RecordKey struct {
	FQDN string
	Zone Zone
	Type RecordType
}

type RecordTarget struct {
	RecordKey

	Address netip.Addr
	TTL     uint32
}

type ZoneResolver interface {
	ResolveZone(ctx context.Context, fqdn string) (Zone, error)
}

type RecordReader[R any] interface {
	Records(ctx context.Context, key RecordKey) ([]Existing[R], error)
}

type RecordWriter[R any] interface {
	Create(ctx context.Context, target RecordTarget) error
	Update(ctx context.Context, target RecordTarget, record R) error
	Delete(ctx context.Context, key RecordKey, record R) error
}

type DDNSClient[R any] interface {
	ZoneResolver
	RecordReader[R]
	RecordWriter[R]
}
