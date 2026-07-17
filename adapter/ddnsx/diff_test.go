package ddnsx

import (
	"context"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

type recordReaderFunc[R any] func(context.Context, RecordKey) ([]Existing[R], error)

func (f recordReaderFunc[R]) Records(ctx context.Context, key RecordKey) ([]Existing[R], error) {
	return f(ctx, key)
}

func TestBuildDiffsSplitsFamiliesAndNormalizesMappedIPv4(t *testing.T) {
	t.Parallel()

	var types []RecordType
	reader := recordReaderFunc[string](func(_ context.Context, key RecordKey) ([]Existing[string], error) {
		types = append(types, key.Type)
		if key.Type == RecordTypeA {
			return []Existing[string]{{Addr: netip.MustParseAddr("192.0.2.1"), Record: "a"}}, nil
		}
		return []Existing[string]{{Addr: netip.MustParseAddr("2001:db8::1"), Record: "aaaa"}}, nil
	})

	diffs, err := BuildDiffs(context.Background(), RecordKey{FQDN: "host.example.com"}, []netip.Addr{
		netip.MustParseAddr("::ffff:192.0.2.1"),
		netip.MustParseAddr("2001:db8::2"),
	}, reader)
	require.NoError(t, err)
	require.Equal(t, []RecordType{RecordTypeA, RecordTypeAAAA}, types)
	require.Equal(t, []Diff[string]{
		{
			Domain: "host.example.com",
			Type:   RecordTypeAAAA,
			Source: netip.MustParseAddr("2001:db8::1"),
			Target: netip.MustParseAddr("2001:db8::2"),
			Action: DDNSActionUpdate,
			Record: "aaaa",
		},
	}, diffs)
}

func TestBuildDiffsEmptyTargetReadsAndDeletesBothFamilies(t *testing.T) {
	t.Parallel()

	reader := recordReaderFunc[int](func(_ context.Context, key RecordKey) ([]Existing[int], error) {
		switch key.Type {
		case RecordTypeA:
			return []Existing[int]{{Addr: netip.MustParseAddr("192.0.2.1"), Record: 4}}, nil
		case RecordTypeAAAA:
			return []Existing[int]{{Addr: netip.MustParseAddr("2001:db8::1"), Record: 6}}, nil
		default:
			return nil, nil
		}
	})

	diffs, err := BuildDiffs(context.Background(), RecordKey{FQDN: "host.example.com"}, nil, reader)
	require.NoError(t, err)
	require.Len(t, diffs, 2)
	require.Equal(t, RecordTypeA, diffs[0].Type)
	require.Equal(t, DDNSActionDelete, diffs[0].Action)
	require.Equal(t, RecordTypeAAAA, diffs[1].Type)
	require.Equal(t, DDNSActionDelete, diffs[1].Action)
}

func TestBuildDiffsRejectsInvalidTarget(t *testing.T) {
	t.Parallel()

	reader := recordReaderFunc[struct{}](func(context.Context, RecordKey) ([]Existing[struct{}], error) {
		t.Fatal("reader should not be called")
		return nil, nil
	})

	_, err := BuildDiffs(context.Background(), RecordKey{}, []netip.Addr{{}}, reader)
	require.EqualError(t, err, "invalid target address")
}

func TestComparePairsTargetsDeterministically(t *testing.T) {
	t.Parallel()

	existing := []Existing[int]{
		{Addr: netip.MustParseAddr("192.0.2.10"), Record: 1},
		{Addr: netip.MustParseAddr("192.0.2.20"), Record: 2},
	}
	target := []netip.Addr{
		netip.MustParseAddr("192.0.2.40"),
		netip.MustParseAddr("192.0.2.30"),
	}

	diffs := Compare("host.example.com", existing, target)
	require.Equal(t, netip.MustParseAddr("192.0.2.30"), diffs[0].Target)
	require.Equal(t, netip.MustParseAddr("192.0.2.40"), diffs[1].Target)
}
