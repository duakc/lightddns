package options

import (
	"github.com/duakc/lightddns/adapter"
	"github.com/duakc/lightddns/infra/badyaml"

	goyaml "github.com/goccy/go-yaml"
	"go.uber.org/zap"
)

type ProviderOption struct {
	AbstractProviderOption `yaml:",inline"`

	Option any `json:"-" yaml:"-"`
}

type _ProviderOption ProviderOption

func (po *ProviderOption) UnmarshalYAML(bs []byte) error {
	err := goyaml.Unmarshal(bs, (*_ProviderOption)(po))
	if err != nil {
		return err
	}
	po.Option, err = adapter.ProviderRegister.CreateOption(po.Type)
	if err != nil {
		return err
	}
	return badyaml.Unmarshal(bs, po.Option)
}

type AbstractProviderOption struct {
	Type string `json:"type"           yaml:"type"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

func (AbstractProviderOption) UsedType() string {
	return "abstract_provider"
}

func (o AbstractProviderOption) CreateLogger(logger *zap.Logger) *zap.Logger {
	return logger.With(
		zap.String("type", "provider"),
		zap.String("provider_type", o.Type)).
		Named(o.Name)
}
