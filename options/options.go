package options

type Options struct {
	Log         OptionLog          `yaml:"log"`
	DataSources []OptionDataSource `yaml:"datasource"`
	Providers   []OptionProvider   `yaml:"provider"`
	Domains     []OptionDomain     `yaml:"domain"`
}
