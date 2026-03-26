package options

type OptionLog struct {
	Level    string `yaml:"level"`
	Disabled bool   `yaml:"disabled"`
	Output   string `yaml:"output"`
}
