package badyaml

import (
	"fmt"
	"strconv"

	"github.com/duakc/mt"
)

type StringOrNumber struct {
	Num int64
	Str string
}

func (bn *StringOrNumber) UnmarshalYAML(data []byte) error {
	unquoted := mt.UnquoteString(string(data))
	*bn = StringOrNumber{} // reset
	if num, err := strconv.ParseInt(unquoted, 10, 64); err == nil {
		bn.Num = num
	} else {
		bn.Str = unquoted
	}
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
