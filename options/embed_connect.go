package options

import (
	"github.com/duakc/lightddns/infra/badyaml"
	"github.com/duakc/lightddns/infra/netool/dialerx"
)

type ConnectOption struct {
	Fwmark uint `json:"fwmark,omitempty" yaml:"fwmark,omitempty"`

	BindAddress4 string `json:"bindAddress4,omitempty" yaml:"bindAddress4,omitempty"`
	BindAddress6 string `json:"bindAddress6,omitempty" yaml:"bindAddress6,omitempty"`

	DialStrategy dialerx.DialStrategy `json:"dialStrategy,omitempty" yaml:"dialStrategy,omitempty"`

	BindInterface badyaml.StringOrNumber `json:"bindInterface,omitempty" yaml:"bindInterface,omitempty"`
}
