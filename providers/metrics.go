package providers

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/metrics"
)

func init() {
	labels := []string{constpkg.MetricLabelName, constpkg.MetricLabelType, constpkg.MetricLabelOperation}

	metrics.RegisterCounter(constpkg.MetricProviderRequestTotal,
		"Total number of provider API requests initiated.", labels)
	metrics.RegisterCounter(constpkg.MetricProviderRequestFailureTotal,
		"Total number of provider API requests that returned an error.", labels)
	metrics.RegisterHistogram(constpkg.MetricProviderRequestDurationSeconds,
		"Duration of a provider API request.", labels, nil)
}
