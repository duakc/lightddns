package options

import "github.com/duakc/lightddns/infra/badyaml"

type matchV4OrV6 struct {
	V4 string `yaml:"v4"`
	V6 string `yaml:"v6"`
}

type HTTPDatasourceOption struct {
	AbstractDatasourceOption `yaml:",inline"`
	ConnectOption            `yaml:",inline"`
	HTTPOption               `yaml:",inline"`

	MatchJson  badyaml.StringOrObject[matchV4OrV6] `yaml:"match-json"`
	MatchRegex badyaml.StringOrObject[matchV4OrV6] `yaml:"match-regex4"`

	Url     badyaml.URL        `yaml:"url"`
	Method  badyaml.HTTPMethod `yaml:"method"`
	Headers badyaml.HTTPHeader `yaml:"headers"`
}
