package options

type Options struct {
	Log         LogOption          `json:"log,omitempty" yaml:"log,omitempty"`
	Datasources []DatasourceOption `json:"datasources"   yaml:"datasources"`
	Providers   []ProviderOption   `json:"providers"     yaml:"providers"`
	Domains     []DomainOption     `json:"domains"       yaml:"domains"`
	Services    []ServiceOption    `json:"services"      yaml:"services"`
}

type VariantOption interface {
	UsedType() string
}
