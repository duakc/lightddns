package lightddns

import (
	"fmt"
	"testing"

	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt"

	"github.com/stretchr/testify/assert"
)

type fakeDatasourceGroup struct {
	group []string
}

func (f fakeDatasourceGroup) Group() []string {
	return f.group
}

func TestResortDatasource(t *testing.T) {
	newDatasourceOption := func(name string, deps []string) options.DatasourceOption {
		opt := options.DatasourceOption{
			AbstractDatasourceOption: options.AbstractDatasourceOption{
				Type: "test",
				Name: name,
			}, Option: new(int),
		} // use new(int) as a fake option

		if len(deps) > 0 {
			opt.Option = &fakeDatasourceGroup{group: deps}
		}
		return opt
	}

	type Case struct {
		Input  []options.DatasourceOption
		Output []string
		Err    string
	}
	cases := []Case{
		{
			Input: []options.DatasourceOption{
				newDatasourceOption("ds1", nil),
				newDatasourceOption("ds2", nil),
				newDatasourceOption("ds3", nil),
			},
			// keep the raw order
			Output: []string{"ds1", "ds2", "ds3"},
		},
		{
			Input: []options.DatasourceOption{
				newDatasourceOption("ds1", []string{"ds2"}),
				newDatasourceOption("ds2", []string{"ds3"}),
				newDatasourceOption("ds3", nil),
			},
			Output: []string{"ds3", "ds2", "ds1"},
		},
		{
			Input: []options.DatasourceOption{
				newDatasourceOption("ds1", []string{"ds2"}),
				newDatasourceOption("ds2", nil),
				newDatasourceOption("ds3", []string{"ds1", "ds2"}),
			},
			Output: []string{"ds2", "ds1", "ds3"},
		},
		{
			Input: []options.DatasourceOption{
				// a datasource depend on itself
				newDatasourceOption("ds1", []string{"ds1"}),
			},
			Err: "circular dependency detected among datasources",
		},
		{
			Input: []options.DatasourceOption{
				// loop
				newDatasourceOption("ds1", []string{"ds2"}),
				newDatasourceOption("ds2", []string{"ds3"}),
				newDatasourceOption("ds3", []string{"ds1"}),
			},
			Err: "circular dependency detected among datasources",
		},
		{
			Input: []options.DatasourceOption{
				// a datasource depend on a non-existed datasource
				newDatasourceOption("ds1", []string{"non-existed"}),
			},
			Err: fmt.Sprintf("datasource(%s) depends on unknown datasource: %q", "ds1", "non-existed"),
		},
		{
			Input: []options.DatasourceOption{
				newDatasourceOption("ds1", nil),
				newDatasourceOption("ds2", []string{"ds1"}),
				newDatasourceOption("ds3", []string{"ds1"}),
			},
			Output: []string{"ds1", "ds2", "ds3"},
		},
	}
	for i := 0; i < len(cases); i++ {
		c := cases[i]
		resorted, err := resortDatasources(c.Input)
		if len(c.Err) != 0 {
			assert.NotNilf(t, err, "testCase.index=%d", i)
			assert.EqualErrorf(t, err, c.Err, "testCase.index=%d", i)
		} else if len(c.Output) > 0 {
			assert.Equalf(t, c.Output, mt.Map(resorted, func(s options.DatasourceOption) string {
				return s.Name
			}), "testCase.index=%d", i)
		}
	}
}
