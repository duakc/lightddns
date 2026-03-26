package options

import "github.com/goccy/go-yaml"

type OptionDataSource struct {
	AbstractProviderOption `yaml:",inline"`

	Option any             `yaml:"-"`
	Raw    yaml.RawMessage `yaml:"-"`
}

type AbstractDatasourceOption struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}
