package ddnsx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type zoneSearchFunc func(context.Context, ZoneName) ([]Zone, error)

func (f zoneSearchFunc) SearchZones(ctx context.Context, keyword ZoneName) ([]Zone, error) {
	return f(ctx, keyword)
}

func TestZoneCacheResolveNormalizesAndUsesMostSpecificZone(t *testing.T) {
	t.Parallel()

	var keywords []ZoneName
	searcher := zoneSearchFunc(func(_ context.Context, keyword ZoneName) ([]Zone, error) {
		keywords = append(keywords, keyword)
		return []Zone{
			{Name: "EXAMPLE.COM.", ID: "parent"},
			{Name: "Sub.Example.com", ID: "child"},
		}, nil
	})

	var cache ZoneCache
	zone, err := cache.Resolve(context.Background(), "Host.Sub.Example.COM.", searcher)
	require.NoError(t, err)
	require.Equal(t, Zone{Name: "sub.example.com", ID: "child"}, zone)
	require.Equal(t, []ZoneName{"host.sub.example.com"}, keywords)

	zone, err = cache.Resolve(context.Background(), "other.sub.example.com", searcher)
	require.NoError(t, err)
	require.Equal(t, Zone{Name: "sub.example.com", ID: "child"}, zone)
	require.Len(t, keywords, 1)
}

func TestZoneCacheResolvePropagatesSearcherError(t *testing.T) {
	t.Parallel()

	upstreamErr := errors.New("upstream unavailable")
	searcher := zoneSearchFunc(func(context.Context, ZoneName) ([]Zone, error) {
		return nil, upstreamErr
	})

	var cache ZoneCache
	_, err := cache.Resolve(context.Background(), "host.example.com", searcher)
	require.ErrorIs(t, err, upstreamErr)
}

func TestZoneCacheResolveReturnsNotFound(t *testing.T) {
	t.Parallel()

	searcher := zoneSearchFunc(func(context.Context, ZoneName) ([]Zone, error) {
		return nil, nil
	})

	var cache ZoneCache
	_, err := cache.Resolve(context.Background(), "host.example.com", searcher)
	require.ErrorIs(t, err, ErrZoneNotFound)
}

func TestNormalizeFQDNRejectsInvalidName(t *testing.T) {
	t.Parallel()

	_, err := NormalizeFQDN("bad..example.com")
	require.ErrorIs(t, err, ErrInvalidDomainName)
}
