package options

import (
	constpkg "github.com/duakc/lightddns/constant"
	"github.com/duakc/lightddns/infra/badyaml"
)

type CommandOutput string

const (
	CommandOutputNone   CommandOutput = "none"
	CommandOutputStdout CommandOutput = "stdout"
	CommandOutputStderr CommandOutput = "stderr"
	CommandOutputAll    CommandOutput = "all"
)

type CommandDatasourceOption struct {
	AbstractDatasourceOption `yaml:",inline"`

	Cmd badyaml.Listable[string] `json:"cmd" yaml:"cmd"`

	ExitCode int `json:"exitCode,omitempty" yaml:"exitCode,omitempty"`

	Env     []string      `json:"env,omitempty"     yaml:"env,omitempty"`
	Output  CommandOutput `json:"output,omitempty"  yaml:"output,omitempty"`
	Capture CommandOutput `json:"capture,omitempty" yaml:"capture,omitempty"`

	// Stdin's priority is higher than StdinContent.
	Stdin        string `json:"stdin,omitempty"        yaml:"stdin,omitempty"`
	StdinContent string `json:"stdinContent,omitempty" yaml:"stdinContent,omitempty"`

	Sync bool `json:"sync,omitempty" yaml:"sync,omitempty"`

	WorkDir string `json:"workDir,omitempty" yaml:"workDir,omitempty"`

	Match MatchOption `json:"match,omitempty" yaml:"match,omitempty"`
}

func (CommandDatasourceOption) UsedType() string {
	return constpkg.DatasourceTypeCommand
}
