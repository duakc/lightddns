package providerx

import (
	"context"
	"errors"
	"testing"

	"github.com/duakc/lightddns/adapter/ddnsx"
	"github.com/duakc/lightddns/adapter/metricx"
	"github.com/duakc/lightddns/infra/metrics"

	"github.com/stretchr/testify/require"
)

type metricsTestClient struct {
	resolveErr error
}

func (c metricsTestClient) ResolveZone(context.Context, string) (ddnsx.Zone, error) {
	return ddnsx.Zone{}, c.resolveErr
}

func (metricsTestClient) Records(context.Context, ddnsx.RecordKey) ([]ddnsx.Existing[int], error) {
	return nil, nil
}

func (metricsTestClient) Create(context.Context, ddnsx.RecordTarget) error {
	return nil
}

func (metricsTestClient) Update(context.Context, ddnsx.RecordTarget, int) error {
	return nil
}

func (metricsTestClient) Delete(context.Context, ddnsx.RecordKey, int) error {
	return nil
}

func TestMetricsClientRecordsLogicalOperation(t *testing.T) {
	t.Parallel()

	registry := metrics.New(true)
	client := NewMetricsClient[int](
		metricx.NewProviderFactory(registry),
		"primary", "test", metricsTestClient{resolveErr: errors.New("unavailable")},
	)

	_, err := client.ResolveZone(context.Background(), "host.example.com")
	require.EqualError(t, err, "unavailable")

	families, err := registry.Gatherer().Gather()
	require.NoError(t, err)
	values := make(map[string]float64, len(families))
	for _, family := range families {
		if len(family.Metric) == 0 {
			continue
		}
		for _, metric := range family.Metric {
			switch family.GetType().String() {
			case "COUNTER":
				values[family.GetName()] += metric.GetCounter().GetValue()
			case "HISTOGRAM":
				values[family.GetName()] += float64(metric.GetHistogram().GetSampleCount())
			}
		}
	}

	require.Equal(t, float64(1), values["lightddns_provider_operation_total"])
	require.Equal(t, float64(1), values["lightddns_provider_operation_failure_total"])
	require.Equal(t, float64(1), values["lightddns_provider_operation_duration_seconds"])
}
