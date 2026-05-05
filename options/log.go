package options

type LogOption struct {
	// @Required @Values trace,debug,info,warn,error,dpanic,panic,fatal
	// @LANG.EN_US Set the log level.
	// @LANG.ZH_CN 设置日志等级
	Level    string `yaml:"level"`
	Disabled bool   `yaml:"disabled"`
	Output   string `yaml:"output"`
}
