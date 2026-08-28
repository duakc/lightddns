package ddnsx

import (
	"context"
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

type RecordSpec struct {
	RecordKey

	TTL uint32
}

// DDNSRecordComparable is implemented by a provider's complete comparison
// object. The object may contain the provider record and any fields that must
// participate in reconciliation.
type DDNSRecordComparable[T comparable] interface {
	comparable
	Compare(T) int
}

type ZoneResolver interface {
	ResolveZone(ctx context.Context, fqdn string) (Zone, error)
}

type RecordReader[T DDNSRecordComparable[T]] interface {
	Records(ctx context.Context, key RecordKey) ([]T, error)
}

type RecordWriter[T DDNSRecordComparable[T]] interface {
	Create(ctx context.Context, target RecordSpec, desired T) error
	Update(ctx context.Context, target RecordSpec, desired T, existing T) error
	Delete(ctx context.Context, key RecordKey, existing T) error
}

type DDNSClient[T DDNSRecordComparable[T]] interface {
	ZoneResolver
	RecordReader[T]
	RecordWriter[T]
}
