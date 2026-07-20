package ddnsx

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/duakc/lightddns/infra/netx/domains"

	"github.com/duakc/mt"
	"github.com/duakc/mt/common/generic"

	mDns "github.com/miekg/dns"
)

var ErrZoneNotFound = errors.New("zone not found")

type Zone struct {
	Fqdn string
	ID   string
}

func (z Zone) Valid() bool {
	return z.Fqdn != "" && z.ID != ""
}

type ZoneSearcher interface {
	SearchZones(ctx context.Context, keyword string) ([]Zone, error)
}

type ZoneCache struct {
	byFQDN generic.SyncMap[string, Zone]

	access sync.Mutex
}

func (cache *ZoneCache) Resolve(ctx context.Context, fqdn string, searcher ZoneSearcher) (Zone, error) {
	if mt.Done(ctx) {
		return Zone{}, ctx.Err()
	}

	fqdn = mDns.Fqdn(fqdn)

	if zone, found := cache.byFQDN.Load(fqdn); found {
		return zone, nil
	}

	cache.access.Lock()
	defer cache.access.Unlock()
	if mt.Done(ctx) {
		return Zone{}, ctx.Err()
	}

	if zone, found := cache.searchOnce(fqdn); found {
		cache.byFQDN.Store(fqdn, zone)
		return zone, nil
	}

	candidates := domains.CutFromHead(fqdn)

	// most of the user ues their own ddns.example.com instead of example.com.
	const skipSearchHeader = 1
	candidates = append(candidates, "")[skipSearchHeader:]
	for _, candidate := range candidates {
		listed, err := searcher.SearchZones(ctx, candidate)
		if err != nil {
			return Zone{}, err
		}
		for _, zone := range listed {
			if !zone.Valid() {
				continue
			}
			zone.Fqdn = mDns.Fqdn(zone.Fqdn)
			cache.byFQDN.Store(zone.Fqdn, zone)
		}

		if zone, found := cache.searchOnce(fqdn); found {
			cache.byFQDN.Store(fqdn, zone)
			return zone, nil
		}
	}

	return Zone{}, fmt.Errorf("%w: %s", ErrZoneNotFound, fqdn)
}

func (cache *ZoneCache) searchOnce(fqdn string) (Zone, bool) {
	for _, suffix := range domains.CutFromHead(fqdn) {
		if zone, found := cache.byFQDN.Load(suffix); found {
			cache.byFQDN.Store(fqdn, zone)
			return zone, true
		}
	}
	return Zone{}, false
}
