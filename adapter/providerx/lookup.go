package providerx

import "github.com/duakc/lightddns/adapter"

func Lookup(manager adapter.ProviderManager, names ...string) ([]adapter.Provider, error) {
	lookupAll, err := manager.LookupAll(names...)
	if err != nil {
		return nil, &ProviderNotFoundError{Err: err}
	}
	return lookupAll, nil
}
