package badyaml

import (
	"fmt"
)

type StringOrNumber struct {
	Num int64
	Str string
}

func (bn *StringOrNumber) UnmarshalYAML(data []byte) error {
	*bn = StringOrNumber{} // reset
	if num, err := UnmarshalType[int64](data); err == nil {
		bn.Num = num
		return nil
	}
	str, err := UnmarshalType[string](data)
	if err != nil {
		return err
	}
	bn.Str = str
	return nil
}

type StringOrObject[T any] struct {
	Str string
	Obj T
}

func (s *StringOrObject[T]) IsString() bool {
	return len(s.Str) > 0
}

func (s *StringOrObject[T]) UnmarshalYAML(data []byte) error {
	if err := Unmarshal(data, &s.Obj); err == nil {
		return nil
	}
	if err := Unmarshal(data, &s.Str); err == nil {
		return nil
	}
	return fmt.Errorf("unknown YAML object: %s", string(data))
}
