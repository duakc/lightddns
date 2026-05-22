package options

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
)

type CommandDatasourceOption struct {
	AbstractDatasourceOption `yaml:",inline"`

	Cmd DualStack[badyaml.Listable[string]] `json:"cmd" yaml:"cmd"`

	Shell    string `json:"shell,omitempty"    yaml:"shell,omitempty"`
	ExitCode int    `json:"exitCode,omitempty" yaml:"exitCode,omitempty"`

	Env badyaml.EnvironmentVariable `json:"env,omitempty" yaml:"env,omitempty"`

	Stdin  string `json:"stdin,omitempty"  yaml:"stdin,omitempty"`
	Stdout string `json:"stdout,omitempty" yaml:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty" yaml:"stderr,omitempty"`
}

func (CommandDatasourceOption) UsedType() string {
	return constpkg.DatasourceTypeCommand
}
