// Package ddnsx provides shared building blocks for DDNS providers.
//
// The diff engine here is provider-agnostic: it computes the minimal set of
// Create / Update / Delete actions needed to reconcile a set of remote DNS
// records with a target list of IP addresses. Providers wire in a fetch
// callback that returns their native record type, and receive Diff[R] values
// that carry that record back unchanged for use in update/delete API calls.
package ddnsx

import (
	"context"
	"fmt"
	"net/netip"

	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/netool"

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
	for _, e := range existing {
		if targetSet[e.Addr] {
			delete(targetSet, e.Addr)
			continue
		}
		unmatched = append(unmatched, e)
	}

	leftover := make([]netip.Addr, 0, len(targetSet))
	for ip := range targetSet {
		leftover = append(leftover, ip)
	}

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

type FetchFunc[R any] func(ctx context.Context, domain string, dnsType string) ([]Existing[R], error)

func BuildDiffs[R any](ctx context.Context, domain string, target []netip.Addr, fetch FetchFunc[R]) ([]Diff[R], error) {
	ipv4, ipv6 := netool.SplitIPv4AndIPv6(target)

	var diffs []Diff[R]
	if len(ipv4) > 0 || len(target) == 0 {
		part, err := fetchAndCompare(ctx, domain, ipv4, constpkg.DNSTypeA, fetch)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, part...)
	}
	if len(ipv6) > 0 || len(target) == 0 {
		part, err := fetchAndCompare(ctx, domain, ipv6, constpkg.DNSTypeAAAA, fetch)
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, part...)
	}
	return diffs, nil
}

func fetchAndCompare[R any](ctx context.Context, domain string, target []netip.Addr, dnsType string, fetch FetchFunc[R]) ([]Diff[R], error) {
	existing, err := fetch(ctx, domain, dnsType)
	if err != nil {
		return nil, err
	}
	return Compare(domain, existing, target), nil
}
