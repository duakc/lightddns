package badyaml

import (
	"io"

	goyaml "github.com/goccy/go-yaml"
)

func Unmarshal(data []byte, v any, options ...goyaml.DecodeOption) error {
	return goyaml.UnmarshalWithOptions(data, v, append(options,
		goyaml.DisallowUnknownField())...)
}

func NewDecoder(r io.Reader, options ...goyaml.DecodeOption) *goyaml.Decoder {
	return goyaml.NewDecoder(r, append(options,
		goyaml.DisallowUnknownField())...)
}

func UnmarshalType[T any](data []byte, options ...goyaml.DecodeOption) (T, error) {
	var v T
	err := Unmarshal(data, &v, options...)
	return v, err
}
