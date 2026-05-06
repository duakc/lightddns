package options

import (
	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/badyaml"

	goyaml "github.com/goccy/go-yaml"
)

type DatasourceOption struct {
	AbstractDatasourceOption

	Option any `json:"-" yaml:"-"`
}

type _DatasourceOption DatasourceOption

func (O *DatasourceOption) UnmarshalYAML(bs []byte) error {
	err := badyaml.Unmarshal(bs, (*_DatasourceOption)(O))
	if err != nil {
		return err
	}
	O.Option, err = adapter.DatasourceRegister.CreateOption(O.Type)
	if err != nil {
		return err
	}
	return goyaml.Unmarshal(bs, O.Option)
}

type AbstractDatasourceOption struct {
	Type string `json:"type" yaml:"type"`
	Name string `json:"name" yaml:"name"`
}

func (AbstractDatasourceOption) UsedType() string {
	return "abstract_datasource"
}
