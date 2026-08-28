// Package ddnsx provides shared building blocks for DDNS providers.
package ddnsx

import (
	"context"
	"fmt"
	"net/netip"
	"slices"

	"github.com/duakc/lightddns/infra/netx"
)

type DDNSAction uint8

const (
	DDNSActionCreate DDNSAction = iota + 1
	DDNSActionUpdate
	DDNSActionDelete
)

func (a DDNSAction) String() string {
	switch a {
	case DDNSActionCreate:
		return "create"
	case DDNSActionUpdate:
		return "update"
	case DDNSActionDelete:
		return "delete"
	}
	return fmt.Sprintf("DDNSAction(%d)", uint8(a))
}

type Diff[T DDNSRecordComparable[T]] struct {
	Domain string
	Type   RecordType
	Action DDNSAction

	Source T
	Target T
}

// Compare returns the operations needed to make existing equal target. The
// provider comparison object is the only equality definition; == is never
// used for matching records.
func Compare[T DDNSRecordComparable[T]](domain string, existing []T, target []T) []Diff[T] {
	if len(existing) == 0 && len(target) == 0 {
		return nil
	}

	old := slices.Clone(existing)
	slices.SortStableFunc(old, func(a, b T) int {
		return a.Compare(b)
	})

	want := slices.Clone(target)
	slices.SortStableFunc(want, func(a, b T) int {
		return a.Compare(b)
	})

	// Remove records that already compare equal. The remaining values are
	// deterministic because both sides are sorted by the provider comparator.
	var unmatchedOld []T
	var unmatchedWant []T
	for i, j := 0, 0; i < len(old) || j < len(want); {
		switch {
		case i == len(old):
			unmatchedWant = append(unmatchedWant, want[j:]...)
			j = len(want)
		case j == len(want):
			for ; i < len(old); i++ {
				unmatchedOld = append(unmatchedOld, old[i])
			}
		case old[i].Compare(want[j]) == 0:
			i++
			j++
		case old[i].Compare(want[j]) < 0:
			unmatchedOld = append(unmatchedOld, old[i])
			i++
		default:
			unmatchedWant = append(unmatchedWant, want[j])
			j++
		}
	}

	diffs := make([]Diff[T], 0, len(unmatchedOld)+len(unmatchedWant))
	pair := min(len(unmatchedOld), len(unmatchedWant))
	for i := range pair {
		diffs = append(diffs, Diff[T]{
			Domain: domain,
			Source: unmatchedOld[i],
			Target: unmatchedWant[i],
			Action: DDNSActionUpdate,
		})
	}
	for i := pair; i < len(unmatchedOld); i++ {
		diffs = append(diffs, Diff[T]{
			Domain: domain,
			Source: unmatchedOld[i],
			Action: DDNSActionDelete,
		})
	}
	for i := pair; i < len(unmatchedWant); i++ {
		diffs = append(diffs, Diff[T]{
			Domain: domain,
			Target: unmatchedWant[i],
			Action: DDNSActionCreate,
		})
	}
	return diffs
}

func BuildDiffs[T DDNSRecordComparable[T]](ctx context.Context, key RecordKey,
	target []netip.Addr, ttl uint32, reader RecordReader[T],
	buildTarget func(netip.Addr, uint32) T,
) ([]Diff[T], error) {
	normalized := make([]netip.Addr, len(target))
	for i, addr := range target {
		if !addr.IsValid() {
			return nil, fmt.Errorf("invalid target address")
		}
		normalized[i] = addr.Unmap()
	}

	ipv4, ipv6 := netx.SplitIPv4AndIPv6(normalized)
	var diffs []Diff[T]
	if len(ipv4) > 0 || len(target) == 0 {
		key.Type = RecordTypeA
		part, err := fetchAndCompare(ctx, key, ipv4, ttl, reader, buildTarget)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, part...)
	}
	if len(ipv6) > 0 || len(target) == 0 {
		key.Type = RecordTypeAAAA
		part, err := fetchAndCompare(ctx, key, ipv6, ttl, reader, buildTarget)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, part...)
	}
	return diffs, nil
}

func fetchAndCompare[T DDNSRecordComparable[T]](ctx context.Context, key RecordKey,
	target []netip.Addr, ttl uint32, reader RecordReader[T],
	buildTarget func(netip.Addr, uint32) T,
) ([]Diff[T], error) {
	existing, err := reader.Records(ctx, key)
	if err != nil {
		return nil, err
	}
	want := make([]T, len(target))
	for i, addr := range target {
		want[i] = buildTarget(addr, ttl)
	}
	diffs := Compare(key.FQDN, existing, want)
	for i := range diffs {
		diffs[i].Type = key.Type
	}
	return diffs, nil
}
