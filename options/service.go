package options

import (
	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/badyaml"

	goyaml "github.com/goccy/go-yaml"
)

type ServiceOption struct {
	AbstractServiceOption `yaml:",inline"`

	Option any `json:"-" yaml:"-"`
}

type _ServiceOption ServiceOption

func (so *ServiceOption) UnmarshalYAML(data []byte) error {
	err := goyaml.Unmarshal(data, (*_ServiceOption)(so))
	if err != nil {
		return err
	}
	so.Option, err = adapter.ServiceRegistry.CreateOption(so.Type)
	if err != nil {
		return err
	}
	return badyaml.Unmarshal(data, so.Option)
}

func (so *ServiceOption) setName(name string) {
	so.AbstractServiceOption.setName(name)
	if setter, canSetName := so.Option.(nameSetter); canSetName {
		setter.setName(name)
	}
}

var (
	_ nameSetter      = (*AbstractServiceOption)(nil)
	_ adapter.Variant = (*AbstractServiceOption)(nil)
)

type AbstractServiceOption struct {
	Type string `json:"type"           yaml:"type"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

func (S *AbstractServiceOption) MajorType() string {
	return "service"
}

func (S *AbstractServiceOption) UsedType() string {
	return "abstract_service"
}

func (S *AbstractServiceOption) setName(name string) { S.Name = name }
func (S *AbstractServiceOption) getName() string     { return S.Name }
