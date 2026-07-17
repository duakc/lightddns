package ddnsx

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

type reconcileClient struct {
	existing   []Existing[int]
	deleteErr  error
	deleteCall int
}

func (c *reconcileClient) ResolveZone(context.Context, string) (Zone, error) {
	return Zone{Name: "example.com"}, nil
}

func (c *reconcileClient) Records(context.Context, RecordKey) ([]Existing[int], error) {
	return c.existing, nil
}

func (c *reconcileClient) Create(context.Context, RecordTarget) error {
	return nil
}

func (c *reconcileClient) Update(context.Context, RecordTarget, int) error {
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
}
