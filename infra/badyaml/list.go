package badyaml

import goyaml "github.com/goccy/go-yaml"

type Listable[T any] struct {
	Value []T `yaml:"-"`
}

func (L *Listable[T]) UnmarshalYAML(bs []byte) error {
	err := goyaml.Unmarshal(bs, &L.Value)
	if err == nil {
		return nil
	}
	var v T
	err = goyaml.Unmarshal(bs, &v)
	if err == nil {
		L.Value = append(L.Value, v)
	}
	return nil
}
