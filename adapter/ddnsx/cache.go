package ddnsx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/duakc/lightddns/infra/netool/domains"

	"github.com/duakc/mt/common/generic"
)

var (
	ErrInvalidDomainName = errors.New("invalid domain name")
	ErrZoneNotFound      = errors.New("zone not found")
)

type ZoneName string

func (n ZoneName) String() string {
	return string(n)
}

type ZoneID string

func (id ZoneID) String() string {
	return string(id)
}

type Zone struct {
	Name ZoneName
	ID   ZoneID
}

func (z Zone) Valid() bool {
	return z.Name != ""
}

type ZoneSearcher interface {
	SearchZones(ctx context.Context, keyword ZoneName) ([]Zone, error)
}

type ZoneCache struct {
	byZoneName generic.SyncMap[ZoneName, Zone]
	byFQDN     generic.SyncMap[string, Zone]

	access sync.Mutex
}

func (cache *ZoneCache) Resolve(ctx context.Context, fqdn string, searcher ZoneSearcher) (Zone, error) {
	fqdn, err := NormalizeFQDN(fqdn)
	if err != nil {
		return Zone{}, err
	}
	if err := ctx.Err(); err != nil {
		return Zone{}, err
	}
	if zone, found := cache.byFQDN.Load(fqdn); found {
		return zone, nil
	}

	cache.access.Lock()
	defer cache.access.Unlock()

	if err := ctx.Err(); err != nil {
		return Zone{}, err
	}
	if zone, found := cache.searchOnce(fqdn); found {
		cache.byFQDN.Store(fqdn, zone)
		return zone, nil
	}

	candidates := domains.CutFromHead(fqdn)
	candidates = append(candidates, "")
	for _, candidate := range candidates {
		listed, err := searcher.SearchZones(ctx, ZoneName(candidate))
		if err != nil {
			return Zone{}, err
		}
		for _, zone := range listed {
			name, err := NormalizeFQDN(zone.Name.String())
			if err != nil {
				continue
			}
			zone.Name = ZoneName(name)
			cache.byZoneName.Store(zone.Name, zone)
		}
		if zone, found := cache.searchOnce(fqdn); found {
			cache.byFQDN.Store(fqdn, zone)
			return zone, nil
		}
	}

	return Zone{}, fmt.Errorf("%w: %s", ErrZoneNotFound, fqdn)
}

func (cache *ZoneCache) searchOnce(fqdn string) (Zone, bool) {
	if zone, ok := cache.byFQDN.Load(fqdn); ok && zone.Valid() {
		return zone, true
	}
	for _, suffix := range domains.CutFromHead(fqdn) {
		if zone, found := cache.byZoneName.Load(ZoneName(suffix)); found {
			return zone, true
		}
	}
	return Zone{}, false
}

func NormalizeFQDN(name string) (string, error) {
	if !domains.IsDomainName(name) {
		return "", fmt.Errorf("%w: %q", ErrInvalidDomainName, name)
	}
	normalized := strings.TrimSuffix(strings.ToLower(name), ".")
	if normalized == "" {
		return "", fmt.Errorf("%w: %q", ErrInvalidDomainName, name)
	}
	return normalized, nil
}
