package metrics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistrySeparatesMetricsBySubsystem(t *testing.T) {
	t.Parallel()

	registry := newDefaultRegistry().(*defaultRegistry)
	first := NewNameFactory(registry, "lightddns", "provider").CounterVec(
		"operation_total", "provider operations", []string{"operation"},
	)
	second := NewNameFactory(registry, "lightddns", "datasource").CounterVec(
		"operation_total", "datasource operations", []string{"name"},
	)

	first.With("list_records").Inc()
	second.With("http").Inc()

	families, err := registry.Gatherer().Gather()
	require.NoError(t, err)
	require.Len(t, families, 2)
	require.Equal(t, "lightddns_datasource_operation_total", families[0].GetName())
	require.Equal(t, "lightddns_provider_operation_total", families[1].GetName())
}

func TestRegistryRejectsDifferentSchemaForSameMetric(t *testing.T) {
	t.Parallel()

	registry := newDefaultRegistry().(*defaultRegistry)
	factory := NewNameFactory(registry, "lightddns", "provider")
	factory.CounterVec("operation_total", "provider operations", []string{"operation"})

	require.PanicsWithValue(t,
		"metrics: lightddns_provider_operation_total registered with a different schema",
		func() {
			factory.CounterVec("operation_total", "different help", []string{"operation"})
		},
	)
}
