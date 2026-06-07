package ddnsx

import (
	"context"
	"sync"

	"github.com/duakc/lightddns/infra/netool/domains"

	"github.com/duakc/mt"
	"github.com/duakc/mt/common/generic"
)

type DomainSearcher interface {
	SearchDomain(ctx context.Context, domain string) map[string]string
}

type DomainIdCache struct {
	domains generic.SyncMap[string, string]

	positive generic.SyncMap[string, string]

	access sync.Mutex
}

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

	fetchResult := fetch.SearchDomain(ctx, domain)
	if fetchResult == nil {
		return ""
	}

	var domainId string

	for k, v := range fetchResult {
		if k == "" || v == "" {
			continue
		}

		if domains.IsSubDomain(domain, k) {
			domainId = v
			cache.positive.Store(domain, v)
		}

		cache.domains.Store(k, v)
	}

	return domainId
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
