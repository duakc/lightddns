package numcompare

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTTL(t *testing.T) {
	require.Equal(t, 0, TTL(0, 600))
	require.Equal(t, 0, TTL(600, 0))
	require.Equal(t, -1, TTL(300, 600))
	require.Equal(t, 1, TTL(600, 300))
	require.Equal(t, 0, TTL(600, 600))
}

func TestBool(t *testing.T) {
	require.Equal(t, -1, Bool(false, true))
	require.Equal(t, 1, Bool(true, false))
	require.Equal(t, 0, Bool(false, false))
	require.Equal(t, 0, Bool(true, true))
}
