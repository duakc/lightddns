package internal

import (
	"fmt"
	urlpkg "net/url"
	"strings"
)

type Params struct {
	m map[string]any
}

func NewParams() *Params {
	return &Params{m: make(map[string]any)}
}

func Add(p *Params, key string, v any) {
	flatten(p.m, key, v)
}

func AddSlice[T any, S ~[]T](p *Params, key string, vs S) {
	for i, v := range vs {
		flatten(p.m, fmt.Sprintf("%s.%d", key, i+1), v)
	}
}

func SetSlice[T any, S ~[]T](p *Params, key string, vs S) {
	prefix := key + "."
	for k := range p.m {
		if k == key || strings.HasPrefix(k, prefix) {
			delete(p.m, k)
		}
	}
	AddSlice(p, key, vs)
}

func (p *Params) Query() urlpkg.Values {
	v := make(urlpkg.Values, len(p.m))
	for k, val := range p.m {
		v.Set(k, fmt.Sprintf("%v", val))
	}
	return v
}

func flatten(out map[string]any, key string, v any) {
	if v == nil {
		return
	}
	switch x := v.(type) {
	case []any:
		for i, item := range x {
			flatten(out, fmt.Sprintf("%s.%d", key, i+1), item)
		}
	case map[string]any:
		for sk, sv := range x {
			flatten(out, fmt.Sprintf("%s.%s", key, sk), sv)
		}
	case []byte:
		out[key] = string(x)
	default:
		out[key] = v
	}
}
