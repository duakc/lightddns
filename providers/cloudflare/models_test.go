package cloudflare

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComparedRecordIgnoresProviderDefaultTTL(t *testing.T) {
	addr := netip.MustParseAddr("192.0.2.1")
	local := ComparedRecord{Addr: addr, TTL: 0, Proxied: true}
	remote := ComparedRecord{Addr: addr, TTL: 600, Proxied: true}
	require.Equal(t, 0, local.Compare(remote))
	require.Equal(t, 0, remote.Compare(local))

	require.NotEqual(t, 0, local.Compare(ComparedRecord{Addr: addr, TTL: 600}))
}
