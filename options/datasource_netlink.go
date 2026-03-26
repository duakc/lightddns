package options

type OptionDataSourceNetlink struct {
	AbstractProviderOption `yaml:",inline"`

	Interface string `yaml:"interface"`
	Index     int    `yaml:"index"`
}
