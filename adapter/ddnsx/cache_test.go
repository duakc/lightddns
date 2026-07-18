package ddnsx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type zoneSearchFunc func(context.Context, string) ([]Zone, error)

func (f zoneSearchFunc) SearchZones(ctx context.Context, keyword string) ([]Zone, error) {
	return f(ctx, keyword)
}

func TestZoneCacheResolveCachesMostSpecificAbsoluteZone(t *testing.T) {
	t.Parallel()

	var keywords []string
	searcher := zoneSearchFunc(func(_ context.Context, keyword string) ([]Zone, error) {
		keywords = append(keywords, keyword)
		return []Zone{
			{Fqdn: "example.com", ID: "parent"},
			{Fqdn: "sub.example.com", ID: "child"},
		}, nil
	})

	var cache ZoneCache
	zone, err := cache.Resolve(context.Background(), "host.sub.example.com", searcher)
	require.NoError(t, err)
	require.Equal(t, Zone{Fqdn: "sub.example.com.", ID: "child"}, zone)
	require.Equal(t, []string{"sub.example.com."}, keywords)

	zone, err = cache.Resolve(context.Background(), "other.sub.example.com", searcher)
	require.NoError(t, err)
	require.Equal(t, Zone{Fqdn: "sub.example.com.", ID: "child"}, zone)
	require.Len(t, keywords, 1)
}

func TestZoneCacheResolvePropagatesSearcherError(t *testing.T) {
	t.Parallel()

	upstreamErr := errors.New("upstream unavailable")
	searcher := zoneSearchFunc(func(context.Context, string) ([]Zone, error) {
		return nil, upstreamErr
	})

	var cache ZoneCache
	_, err := cache.Resolve(context.Background(), "host.example.com", searcher)
	require.ErrorIs(t, err, upstreamErr)
}

func TestZoneCacheResolveReturnsNotFound(t *testing.T) {
	t.Parallel()

	searcher := zoneSearchFunc(func(context.Context, string) ([]Zone, error) {
		return nil, nil
	})

	var cache ZoneCache
	_, err := cache.Resolve(context.Background(), "host.example.com", searcher)
	require.ErrorIs(t, err, ErrZoneNotFound)
}
