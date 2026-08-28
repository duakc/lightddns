package ddnsx

import (
	"context"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

type testComparedRecord struct {
	Addr   netip.Addr
	TTL    uint32
	Record int
}

func (r testComparedRecord) Address() netip.Addr { return r.Addr }

func (r testComparedRecord) Compare(other testComparedRecord) int {
	if c := r.Addr.Compare(other.Addr); c != 0 {
		return c
	}
	if r.TTL < other.TTL {
		return -1
	}
	if r.TTL > other.TTL {
		return 1
	}
	return 0
}

type recordReaderFunc[T DDNSRecordComparable[T]] func(context.Context, RecordKey) ([]T, error)

func (f recordReaderFunc[T]) Records(ctx context.Context, key RecordKey) ([]T, error) {
	return f(ctx, key)
}

func TestBuildDiffsSplitsFamiliesAndNormalizesMappedIPv4(t *testing.T) {
	t.Parallel()

	var types []RecordType
	reader := recordReaderFunc[testComparedRecord](func(_ context.Context, key RecordKey) ([]testComparedRecord, error) {
		types = append(types, key.Type)
		if key.Type == RecordTypeA {
			return []testComparedRecord{{
				Addr: netip.MustParseAddr("192.0.2.1"), TTL: 300, Record: 1,
			}}, nil
		}
		return []testComparedRecord{{
			Addr: netip.MustParseAddr("2001:db8::1"), Record: 2,
		}}, nil
	})

	diffs, err := BuildDiffs(context.Background(), RecordKey{FQDN: "host.example.com"}, []netip.Addr{
		netip.MustParseAddr("::ffff:192.0.2.1"),
		netip.MustParseAddr("2001:db8::2"),
	}, 300, reader, func(addr netip.Addr, ttl uint32) testComparedRecord {
		return testComparedRecord{Addr: addr, TTL: ttl}
	})
	require.NoError(t, err)
	require.Equal(t, []RecordType{RecordTypeA, RecordTypeAAAA}, types)
	require.Len(t, diffs, 1)
	require.Equal(t, RecordTypeAAAA, diffs[0].Type)
	require.Equal(t, DDNSActionUpdate, diffs[0].Action)
	require.Equal(t, netip.MustParseAddr("2001:db8::1"), diffs[0].Source.Addr)
	require.Equal(t, netip.MustParseAddr("2001:db8::2"), diffs[0].Target.Addr)
}

func TestBuildDiffsEmptyTargetReadsAndDeletesBothFamilies(t *testing.T) {
	t.Parallel()

	reader := recordReaderFunc[testComparedRecord](func(_ context.Context, key RecordKey) ([]testComparedRecord, error) {
		switch key.Type {
		case RecordTypeA:
			return []testComparedRecord{{Addr: netip.MustParseAddr("192.0.2.1")}}, nil
		case RecordTypeAAAA:
			return []testComparedRecord{{Addr: netip.MustParseAddr("2001:db8::1")}}, nil
		default:
			return nil, nil
		}
	})

	diffs, err := BuildDiffs(context.Background(), RecordKey{FQDN: "host.example.com"}, nil, 300, reader,
		func(addr netip.Addr, ttl uint32) testComparedRecord {
			return testComparedRecord{Addr: addr, TTL: ttl}
		})
	require.NoError(t, err)
	require.Len(t, diffs, 2)
	require.Equal(t, RecordTypeA, diffs[0].Type)
	require.Equal(t, DDNSActionDelete, diffs[0].Action)
	require.Equal(t, RecordTypeAAAA, diffs[1].Type)
	require.Equal(t, DDNSActionDelete, diffs[1].Action)
}

func TestBuildDiffsRejectsInvalidTarget(t *testing.T) {
	t.Parallel()

	reader := recordReaderFunc[testComparedRecord](func(context.Context, RecordKey) ([]testComparedRecord, error) {
		t.Fatal("reader should not be called")
		return nil, nil
	})

	_, err := BuildDiffs(context.Background(), RecordKey{}, []netip.Addr{{}}, 300, reader,
		func(addr netip.Addr, ttl uint32) testComparedRecord {
			return testComparedRecord{Addr: addr, TTL: ttl}
		})
	require.EqualError(t, err, "invalid target address")
}

func TestComparePairsTargetsDeterministically(t *testing.T) {
	t.Parallel()

	existing := []testComparedRecord{
		{Addr: netip.MustParseAddr("192.0.2.10"), Record: 1},
		{Addr: netip.MustParseAddr("192.0.2.20"), Record: 2},
	}
	target := []testComparedRecord{
		{Addr: netip.MustParseAddr("192.0.2.40")},
		{Addr: netip.MustParseAddr("192.0.2.30")},
	}

	diffs := Compare("host.example.com", existing, target)
	require.Equal(t, netip.MustParseAddr("192.0.2.30"), diffs[0].Target.Addr)
	require.Equal(t, netip.MustParseAddr("192.0.2.40"), diffs[1].Target.Addr)
}

func TestCompareIncludesTTLChanges(t *testing.T) {
	t.Parallel()

	addr := netip.MustParseAddr("192.0.2.10")
	diffs := Compare("host.example.com", []testComparedRecord{
		{Addr: addr, TTL: 60, Record: 1},
	}, []testComparedRecord{{Addr: addr, TTL: 300}})

	require.Len(t, diffs, 1)
	require.Equal(t, DDNSActionUpdate, diffs[0].Action)
	require.Equal(t, uint32(300), diffs[0].Target.TTL)
}
