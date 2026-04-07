package options

type Options struct {
	Log         LogOption          `yaml:"log"`
	DataSources []DatasourceOption `yaml:"datasource"`
	Providers   []ProviderOption   `yaml:"provider"`
	Domains     []DomainOption     `yaml:"domain"`
}
