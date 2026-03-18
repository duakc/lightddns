package options

type DataSource interface {
	Type() string
}
type OptionDataSource struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`

	DataSource DataSource `yaml:"-"`
}
