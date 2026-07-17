package adapter

import "fmt"

func AutoName(variant Variant, index int) string {
	return fmt.Sprintf("%s[%d]", variant.MajorType(), index)
}

type Variant interface {
	MajorType() string
	UsedType() string
}
