package options

type LogOption struct {
	// @Values trace,debug,info,warn,error,dpanic,panic,fatal
	// @LANG.EN_US Set the log level.
	// @LANG.ZH_CN 设置日志等级
	Level    string `json:"level,omitempty"    yaml:"level,omitempty"`
	Disabled bool   `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Output   string `json:"output,omitempty"   yaml:"output,omitempty"`
}
