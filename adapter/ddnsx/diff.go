// Package ddnsx provides shared building blocks for DDNS providers.
package ddnsx

import (
	"context"
	"fmt"
	"net/netip"
	"slices"

	"github.com/duakc/lightddns/infra/netx"

	"github.com/duakc/mt"
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

type Existing[R any] struct {
	Addr   netip.Addr
	Record R
}

type Diff[R any] struct {
	Domain         string
	Type           RecordType
	Source, Target netip.Addr
	Action         DDNSAction
	Record         R
}

func Compare[R any](domain string, existing []Existing[R], target []netip.Addr) []Diff[R] {
	if len(existing) == 0 && len(target) == 0 {
		return nil
	}

	targetSet := mt.Set(target)
	unmatched := make([]Existing[R], 0, len(existing))
	for _, record := range existing {
		if targetSet[record.Addr] {
			delete(targetSet, record.Addr)
			continue
		}
		unmatched = append(unmatched, record)
	}

	leftover := make([]netip.Addr, 0, len(targetSet))
	for ip := range targetSet {
		leftover = append(leftover, ip)
	}
	slices.SortFunc(leftover, netip.Addr.Compare)

	diffs := make([]Diff[R], 0, len(unmatched)+len(leftover))
	pair := min(len(unmatched), len(leftover))
	for i := range pair {
		diffs = append(diffs, Diff[R]{
			Domain: domain,
			Source: unmatched[i].Addr,
			Target: leftover[i],
			Action: DDNSActionUpdate,
			Record: unmatched[i].Record,
		})
	}
	for i := pair; i < len(unmatched); i++ {
		diffs = append(diffs, Diff[R]{
			Domain: domain,
			Source: unmatched[i].Addr,
			Action: DDNSActionDelete,
			Record: unmatched[i].Record,
		})
	}
	for i := pair; i < len(leftover); i++ {
		diffs = append(diffs, Diff[R]{
			Domain: domain,
			Target: leftover[i],
			Action: DDNSActionCreate,
		})
	}
	return diffs
}

func BuildDiffs[R any](ctx context.Context, key RecordKey,
	target []netip.Addr, reader RecordReader[R],
) ([]Diff[R], error) {
	normalized := make([]netip.Addr, len(target))
	for i, addr := range target {
		if !addr.IsValid() {
			return nil, fmt.Errorf("invalid target address")
		}
		normalized[i] = addr.Unmap()
	}

	ipv4, ipv6 := netx.SplitIPv4AndIPv6(normalized)
	var diffs []Diff[R]
	if len(ipv4) > 0 || len(target) == 0 {
		key.Type = RecordTypeA
		part, err := fetchAndCompare(ctx, key, ipv4, reader)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, part...)
	}
	if len(ipv6) > 0 || len(target) == 0 {
		key.Type = RecordTypeAAAA
		part, err := fetchAndCompare(ctx, key, ipv6, reader)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, part...)
	}
	return diffs, nil
}

func fetchAndCompare[R any](ctx context.Context, key RecordKey,
	target []netip.Addr, reader RecordReader[R],
) ([]Diff[R], error) {
	existing, err := reader.Records(ctx, key)
	if err != nil {
		return nil, err
	}
	diffs := Compare(key.FQDN, existing, target)
	for i := range diffs {
		diffs[i].Type = key.Type
	}
	return diffs, nil
}
