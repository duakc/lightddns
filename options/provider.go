package options

type Provider interface {
	Provider() string
}

type OptionProvider struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`

	P Provider `yaml:"-"`
}
