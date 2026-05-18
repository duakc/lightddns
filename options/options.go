package options

type Options struct {
	Log         LogOption          `json:"log,omitempty" yaml:"log,omitempty"`
	Datasources []DatasourceOption `json:"datasources" yaml:"datasources"`
	Providers   []ProviderOption   `json:"providers"     yaml:"providers"`
	Domains     []DomainOption     `json:"domains"       yaml:"domains"`
}

type VariantOption interface {
	UsedType() string
}
