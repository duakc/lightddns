package badyaml

import (
	"fmt"
	"strconv"

	"github.com/duakc/mt"

	goyaml "github.com/goccy/go-yaml"
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

func (s *StringOrObject[T]) UnmarshalYAML(data []byte) error {
	if err := goyaml.Unmarshal(data, &s.Str); err == nil {
		return nil
	}
	if err := goyaml.Unmarshal(data, &s.Obj); err == nil {
		return nil
	}
	return fmt.Errorf("unknown YAML object: %s", string(data))
}
