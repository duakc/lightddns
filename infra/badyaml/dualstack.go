package badyaml

import "errors"

type DualStack[T any] struct {
	IPv4 T `json:"ipv4" yaml:"ipv4"`
	IPv6 T `json:"ipv6" yaml:"ipv6"`
}

type _DualStack[T any] DualStack[T]

func (d *DualStack[T]) UnmarshalYAML(data []byte) error {
	var (
		asObject  _DualStack[T]
		objectErr error
	)

	if asObject, objectErr = UnmarshalType[_DualStack[T]](data); objectErr == nil {
		d.IPv4 = asObject.IPv4
		d.IPv6 = asObject.IPv6
		return nil
	}

	single, err := UnmarshalType[T](data)
	if err != nil {
		return errors.Join(objectErr, err)
	}

	d.IPv4 = single
	d.IPv6 = single
	return nil
}
