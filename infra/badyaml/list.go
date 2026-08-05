package badyaml

import "errors"

type Listable[T any] struct {
	Value []T `yaml:"-"`
}

func (L *Listable[T]) UnmarshalYAML(bs []byte) error {
	err := Unmarshal(bs, &L.Value)
	if err == nil {
		return nil
	}

	var v T
	commonErr := Unmarshal(bs, &v)
	if commonErr == nil {
		L.Value = append([]T{}, v)
		return nil
	}
	return errors.Join(commonErr, err)
}
