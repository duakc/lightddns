package options

type DatasourceIsGroup interface {
	Group() []string
}

type AbstractDatasourceGroupOption struct {
	AbstractDatasourceOption `yaml:",inline"`
	Datasources              []string `yaml:"datasources"`
}

func (x AbstractDatasourceGroupOption) Group() []string {
	return x.Datasources
}
