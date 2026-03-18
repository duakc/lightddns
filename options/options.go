package options

type Options struct {
	Log         OptionLog          `yaml:"log"`
	DataSources []OptionDataSource `yaml:"dataSources"`
	Providers   []OptionProvider   `yaml:"providers"`
	Domains     []OptionDomain     `yaml:"domains"`
}
