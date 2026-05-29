package options

import (
	"fmt"

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
		if cur.Name != "" {
			continue
		}
		cur.setName(autoName(&cur.AbstractDatasourceOption, cur.Type, i))
	}
	for i := range O.Providers {
		cur := &O.Providers[i]
		if cur.Name != "" {
			continue
		}
		cur.setName(autoName(&cur.AbstractProviderOption, cur.Type, i))
	}
	for i := range O.Services {
		cur := &O.Services[i]
		if cur.Name != "" {
			continue
		}
		cur.setName(autoName(&cur.AbstractServiceOption, cur.Type, i))
	}
	return nil
}

func autoName(variant VariantOption, typ string, index int) string {
	return fmt.Sprintf("%s_%s[%d]", variant.MajorType(), typ, index)
}

type nameSetter interface {
	setName(string)
}

type VariantOption interface {
	MajorType() string
	UsedType() string
}
