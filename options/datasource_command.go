package options

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
)

type CommandDatasourceOption struct {
	AbstractDatasourceOption

	Cmd DualStack[string] `json:"cmd" yaml:"cmd"`

	Shell    string `json:"shell,omitempty"    yaml:"shell,omitempty"`
	ExitCode int    `json:"exitCode,omitempty" yaml:"exitCode,omitempty"`

	Env badyaml.EnvironmentVariable `json:"env,omitempty" yaml:"env,omitempty"`
}

func (CommandDatasourceOption) UsedType() string {
	return constpkg.DatasourceTypeCommand
}
