package datasources

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/metrics"
)

func init() {
	labels := []string{constpkg.MetricLabelName, constpkg.MetricLabelType, constpkg.MetricLabelFamily}

	metrics.RegisterCounter(constpkg.MetricDatasourceQueryTotal,
		"Total number of datasource IP lookups initiated.", labels)
	metrics.RegisterCounter(constpkg.MetricDatasourceQueryFailureTotal,
		"Total number of datasource IP lookups that returned an error.", labels)
	metrics.RegisterHistogram(constpkg.MetricDatasourceQueryDurationSeconds,
		"Duration of a datasource IP lookup.", labels, nil)
	metrics.RegisterGauge(constpkg.MetricDatasourceIPCount,
		"Number of IP addresses returned by the most recent lookup.", labels)
}
