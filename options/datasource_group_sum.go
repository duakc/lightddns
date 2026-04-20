package options

type DatasourceGroupSumOption struct {
	AbstractDatasourceGroupOption `yaml:",inline"`

	IgnoreDownstreamError bool `yaml:"ignore-downstream-error"`
}
