package options

import (
	"github.com/duakc/lightddns/infra/badyaml"
)

type MatchOption struct {
	JQ     *badyaml.JQ    `json:"jq,omitempty"    yaml:"jq,omitempty"`
	Regexp *badyaml.Regex `json:"regex,omitempty" yaml:"regex,omitempty"`
}
