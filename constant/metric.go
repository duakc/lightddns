package constant

// Metric label keys. Leaf metric names live at the top of the file that
// declares the metric (see e.g. domain.go, providers/cloudflare/cloudflare.go);
// the ddns namespace and per-variant subsystem are added by
// adapter/ddnsmetric.Factory at registration time.
const (
	MetricLabelDomain    = "domain"
	MetricLabelName      = "name"
	MetricLabelType      = "type"
	MetricLabelFamily    = "family"
	MetricLabelOperation = "operation"
	MetricLabelVersion   = "version"
	MetricLabelBranch    = "branch"
)
