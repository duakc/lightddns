package options

type OptionDataSourceNetlink struct {
	AbstractDatasourceOption `yaml:",inline"`

	Interface    string `yaml:"interface"`
	Index        int    `yaml:"index"`
	AllowPrivate bool   `yaml:"allow-private"`
}
