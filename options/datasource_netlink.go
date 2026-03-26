package options

type OptionDataSourceNetlink struct {
	abstractDatasourceOption `yaml:",inline"`

	Interface string `yaml:"interface"`
	Index     int    `yaml:"index"`
}
