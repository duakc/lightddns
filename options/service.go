package options

import (
	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/badyaml"

	goyaml "github.com/goccy/go-yaml"
)

type AbstractServiceOption struct {
	Type string `json:"type"           yaml:"type"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

type ServiceOption struct {
	AbstractServiceOption `yaml:",inline"`

	Option any `json:"-" yaml:"-"`
}

type _ServiceOption ServiceOption

func (S *ServiceOption) UnmarshalYAML(data []byte) error {
	err := goyaml.Unmarshal(data, (*_ServiceOption)(S))
	if err != nil {
		return err
	}
	S.Option, err = adapter.ServiceRegistry.CreateOption(S.Type)
	if err != nil {
		return err
	}
	return badyaml.Unmarshal(data, S.Option)
}

func (S *ServiceOption) UsedType() string {
	return "abstract_service"
}
