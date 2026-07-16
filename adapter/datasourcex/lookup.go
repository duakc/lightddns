package datasourcex

import "github.com/duakc/lightddns/adapter"

func Lookup(manager adapter.DatasourceManager, names ...string) ([]adapter.Datasource, error) {
	all, err := manager.LookupAll(names...)
	if err != nil {
		return nil, &DatasourceNotFoundError{err}
	}
	return all, nil
}
