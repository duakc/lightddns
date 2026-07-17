package ddnsx

import (
	"context"
	"sync"

	"github.com/duakc/lightddns/infra/netool/domains"

	"github.com/duakc/mt"
	"github.com/duakc/mt/common/generic"
)

// DomainSearcher fetches the user's zones from upstream, optionally
// filtered by a substring keyword.
//
// Contract:
//   - keyword "" means "list everything" — the caller wants a full sweep.
//   - non-empty keyword is a substring / exact-name filter, depending on
//     what the provider's API supports. Providers whose API has no
//     filter (e.g. Cloudflare ListZones) simply ignore it and return
//     the full set.
//   - returns nil on transport / API failure so the cache treats it as
//     "no result" without poisoning the negative path.
//
// The cut-from-head iteration that turns a queried FQDN into a sequence of
// candidate keywords lives in [DomainIdCache.LoadOrStore] — implementers
// of SearchDomain just translate one keyword into one upstream filter.
type DomainSearcher interface {
	SearchDomain(ctx context.Context, keyword string) map[string]string
}

type DomainIdCache struct {
	domains generic.SyncMap[string, string]

	positive generic.SyncMap[string, string]

	access sync.Mutex
}

// LoadOrStore returns the zone id for the parent zone of `domain`,
// fetching from upstream via `fetch` if it isn't already cached.
//
// On a miss we walk suffixes of `domain` longest-first and call
// fetch.SearchDomain with each as the upstream filter:
//
//	SearchDomain(ctx, "a.example.com")
//	SearchDomain(ctx, "example.com")
//	SearchDomain(ctx, "com")
//	SearchDomain(ctx, "")            // final fallback: list everything
//
// Each returned zone is memoised in cache.domains; the loop short-circuits
// as soon as a listed zone is a parent of `domain`. This minimises upstream
// calls for providers whose API supports a keyword filter and degenerates
// gracefully (one call returns everything) for providers that don't.
func (cache *DomainIdCache) LoadOrStore(ctx context.Context,
	domain string, fetch DomainSearcher,
) string {
	if domain == "" || mt.Done(ctx) {
		return ""
	}

	if domainId, found := cache.positive.Load(domain); found {
		return domainId
	}

	cache.access.Lock()
	defer cache.access.Unlock()
	if found := cache.searchOnce(domain); found != "" {
		cache.positive.Store(domain, found)
		return found
	}

	// Most users don't directly use root domain when using DDNS,
	// so we skip one level to speed up the lookup process.
	//
	// For example, most users would use home.example.com instead of example.com.
	const skipHead = 1

	for _, keyword := range append(domains.CutFromHead(domain), "")[skipHead:] {
		listed := fetch.SearchDomain(ctx, keyword)
		if listed == nil {
			return ""
		}
		for k, v := range listed {
			if k == "" || v == "" {
				continue
			}
			cache.domains.Store(k, v)
		}
		if id := cache.searchOnce(domain); id != "" {
			cache.positive.Store(domain, id)
			return id
		}
	}

	return ""
}

func (cache *DomainIdCache) searchOnce(domain string) string {
	if posi, ok := cache.positive.Load(domain); ok && posi != "" {
		return posi
	}

	cuts := domains.CutFromHead(domain)
	for _, domainPart := range cuts {
		if id, foundId := cache.domains.Load(domainPart); foundId {
			return id
		}
	}

	return ""
}
