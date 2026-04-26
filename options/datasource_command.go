package options

import "github.com/duakc/lightddns/infra/badyaml"

type CommandDatasourceOption struct {
	AbstractDatasourceOption `yaml:",inline"`

	// CmdV4 and CmdV6 is a shell command: support pipe, redirector and so on
	// It use the current system shell to execute the command
	CmdV4    string `yaml:"cmd-v4"`
	CmdV6    string `yaml:"cmd-v6"`
	Shell    string `yaml:"shell"`
	ExitCode int    `yaml:"exit-code"`

	Env badyaml.EnvironmentVariable `yaml:"env"`
}
