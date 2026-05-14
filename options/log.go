package options

import "github.com/duakc/lightddns/infra/badyaml"

type LogOption struct {
	Level badyaml.LogLevel `json:"level,omitempty" yaml:"level,omitempty"`

	Disabled bool   `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Output   string `json:"output,omitempty"   yaml:"output,omitempty"`
}
