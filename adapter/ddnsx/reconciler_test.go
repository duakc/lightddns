package ddnsx

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

type reconcileClient struct {
	existing     []Existing[int]
	deleteErr    error
	deleteCall   int
	resolvedFQDN string
	recordFQDN   string
}

func (c *reconcileClient) ResolveZone(_ context.Context, fqdn string) (Zone, error) {
	c.resolvedFQDN = fqdn
	return Zone{Fqdn: "example.com.", ID: "zone-id"}, nil
}

func (c *reconcileClient) Records(_ context.Context, key RecordKey) ([]Existing[int], error) {
	c.recordFQDN = key.FQDN
	return c.existing, nil
}

func (c *reconcileClient) Create(context.Context, RecordSpec) error {
	return nil
}

func (c *reconcileClient) Update(context.Context, RecordSpec, int) error {
	return nil
}

func (c *reconcileClient) Delete(context.Context, RecordKey, int) error {
	c.deleteCall++
	if c.deleteCall == 2 {
		return c.deleteErr
	}
	return nil
}

func TestReconcilerReportsPartialSuccess(t *testing.T) {
	t.Parallel()

	mutationErr := errors.New("second delete failed")
	client := &reconcileClient{
		existing: []Existing[int]{
			{Addr: netip.MustParseAddr("192.0.2.1"), Record: 1},
			{Addr: netip.MustParseAddr("192.0.2.2"), Record: 2},
		},
		deleteErr: mutationErr,
	}

	changed, err := NewReconciler[int](nil, client).Update(
		context.Background(), "Host.Example.COM.", 300, nil,
	)
	require.True(t, changed)
	require.ErrorIs(t, err, mutationErr)
	require.Equal(t, 2, client.deleteCall)
	require.Equal(t, "Host.Example.COM.", client.resolvedFQDN)
	require.Equal(t, "Host.Example.COM.", client.recordFQDN)
}
