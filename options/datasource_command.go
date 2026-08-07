package options

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
)

type CommandOutput string

const (
	CommandOutputStdout CommandOutput = "stdout"
	CommandOutputStderr CommandOutput = "stderr"
	CommandOutputAll    CommandOutput = "all"
)

type CommandDatasourceOption struct {
	AbstractDatasourceOption `yaml:",inline"`

	Cmd badyaml.Listable[string] `json:"cmd" yaml:"cmd"`

	ExitCode int `json:"exitCode,omitempty" yaml:"exitCode,omitempty"`

	Env    []string        `json:"env,omitempty" yaml:"env,omitempty"`
	Output []CommandOutput `json:"output,omitempty" yaml:"output,omitempty"`

	Stdin string `json:"stdin,omitempty"  yaml:"stdin,omitempty"`
	// Stdout string `json:"stdout,omitempty" yaml:"stdout,omitempty"`
	// Stderr string `json:"stderr,omitempty" yaml:"stderr,omitempty"`

	JQ    *badyaml.JQ    `json:"jq,omitempty" yaml:"jq,omitempty"`
	Regex *badyaml.Regex `json:"regex,omitempty" yaml:"regex,omitempty"`
}

func (CommandDatasourceOption) UsedType() string {
	return constpkg.DatasourceTypeCommand
}
