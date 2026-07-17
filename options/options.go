package options

import (
	"github.com/duakc/lightddns/adapter"

	goyaml "github.com/goccy/go-yaml"
)

type Options struct {
	Log         LogOption          `json:"log,omitempty" yaml:"log,omitempty"`
	Datasources []DatasourceOption `json:"datasources"   yaml:"datasources"`
	Providers   []ProviderOption   `json:"providers"     yaml:"providers"`
	Domains     []DomainOption     `json:"domains"       yaml:"domains"`
	Services    []ServiceOption    `json:"services"      yaml:"services"`
}

type _Options Options

func (O *Options) UnmarshalYAML(data []byte) error {
	if err := goyaml.Unmarshal(data, (*_Options)(O)); err != nil {
		return err
	}

	for i := range O.Datasources {
		cur := &O.Datasources[i]
		if cur.Name == "" {
			cur.setName(adapter.AutoName(&cur.AbstractDatasourceOption, i))
		}
	}
	for i := range O.Providers {
		cur := &O.Providers[i]
		if cur.Name == "" {
			cur.setName(adapter.AutoName(&cur.AbstractProviderOption, i))
		}
	}
	for i := range O.Services {
		cur := &O.Services[i]
		if cur.Name == "" {
			cur.setName(adapter.AutoName(&cur.AbstractServiceOption, i))
		}
	}
	return nil
}

type nameSetter interface {
	setName(string)
}
