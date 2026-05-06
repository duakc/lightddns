package options

type DatasourceIsGroup interface {
	Group() []string
}

type AbstractDatasourceGroupOption struct {
	AbstractDatasourceOption

	Datasources []string `json:"datasources" yaml:"datasources"`
}

func (x AbstractDatasourceGroupOption) Group() []string {
	return x.Datasources
}

func (AbstractDatasourceGroupOption) UsedType() string {
	return "abstract_datasource_group"
}
