package prometheus

import (
	"context"
	"testing"

	"github.com/duakc/lightddns/options"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewUsesDefaultPort(t *testing.T) {
	service, err := New(context.Background(), zap.NewNop(), options.PrometheusServiceOption{
		AbstractServiceOption: options.AbstractServiceOption{
			Type: ServiceType,
			Name: "test-prometheus",
		},
		Enabled: true,
	})
	require.NoError(t, err)
	require.Equal(t, ":9001", service.(*Prometheus).addr)
}
