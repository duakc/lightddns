package constant

// Metric names. All metrics start with "ddns_" prefix.
const (
	MetricDomainActivationTotal       = "ddns_domain_activation_total"
	MetricDomainUpdateSuccessTotal    = "ddns_domain_update_success_total"
	MetricDomainUpdateFailureTotal    = "ddns_domain_update_failure_total"
	MetricDomainNoIPAddressTotal      = "ddns_domain_no_ip_address_total"
	MetricDomainActualUpdateTotal     = "ddns_domain_actual_update_total"
	MetricDomainUpdateDurationSeconds = "ddns_domain_update_duration_seconds"
	MetricDomainLastUpdateTimestamp   = "ddns_domain_last_update_timestamp_seconds"

	MetricDatasourceQueryTotal           = "ddns_datasource_query_total"
	MetricDatasourceQueryFailureTotal    = "ddns_datasource_query_failure_total"
	MetricDatasourceQueryDurationSeconds = "ddns_datasource_query_duration_seconds"
	MetricDatasourceIPCount              = "ddns_datasource_ip_count"

	MetricProviderRequestTotal           = "ddns_provider_request_total"
	MetricProviderRequestFailureTotal    = "ddns_provider_request_failure_total"
	MetricProviderRequestDurationSeconds = "ddns_provider_request_duration_seconds"

	MetricBuildInfo = "ddns_build_info"
)

const (
	MetricLabelDomain    = "domain"
	MetricLabelName      = "name"
	MetricLabelType      = "type"
	MetricLabelFamily    = "family"
	MetricLabelOperation = "operation"
	MetricLabelVersion   = "version"
	MetricLabelBranch    = "branch"
)

const (
	MetricFamilyIPv4 = "ipv4"
	MetricFamilyIPv6 = "ipv6"
)
