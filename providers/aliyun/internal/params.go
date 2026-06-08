package internal

import (
	"fmt"
	urlpkg "net/url"
	"strings"
)

// Params collects RPC action parameters before serialising them into a
// url.Values for the V3 canonical request. The underlying storage is
// map[string]any keyed by the dotted, flattened parameter name (e.g.
// "DomainName", "InstanceId.1", "Tags.1.Key").
//
// Values are added via the free functions [Add], [AddSlice], [SetSlice]
// rather than methods, because Go does not yet allow type parameters on
// methods and we want a uniform API surface between scalar and generic
// adders.
type Params struct {
	m map[string]any
}

func NewParams() *Params {
	return &Params{m: make(map[string]any)}
}

// Add records v under key.
//
// Nested values are flattened in place, matching Aliyun's RPC convention:
//   - []any        →  key.1=…, key.2=…  (1-indexed)
//   - map[string]X →  key.subkey=…
//   - []byte       →  string(v)
//   - other        →  stored as-is, stringified at [Params.Query] time
//
// Zero values (empty strings, 0, false) are NOT filtered — only nil is
// dropped. Callers decide whether an unset optional field should be added.
func Add(p *Params, key string, v any) {
	flatten(p.m, key, v)
}

// AddSlice expands a typed slice into 1-indexed dotted keys
// (key.1, key.2, …). Each element is recursively flattened, so a slice of
// maps becomes key.{i}.{subkey} entries.
//
// The dual type parameter (`S ~[]T`) lets callers pass named slice types
// without an explicit conversion.
func AddSlice[T any, S ~[]T](p *Params, key string, vs S) {
	for i, v := range vs {
		flatten(p.m, fmt.Sprintf("%s.%d", key, i+1), v)
	}
}

// SetSlice replaces any existing entries at key or under the "key." prefix
// with the new flattened slice. Use this when the caller wants exact
// replacement semantics for a previously-added entry list.
func SetSlice[T any, S ~[]T](p *Params, key string, vs S) {
	prefix := key + "."
	for k := range p.m {
		if k == key || strings.HasPrefix(k, prefix) {
			delete(p.m, k)
		}
	}
	AddSlice(p, key, vs)
}

// Query returns the accumulated parameters as url.Values, ready to be
// attached to a request's URL and signed via [SigContext]. Values are
// stringified via fmt.Sprintf("%v", …).
func (p *Params) Query() urlpkg.Values {
	v := make(urlpkg.Values, len(p.m))
	for k, val := range p.m {
		v.Set(k, fmt.Sprintf("%v", val))
	}
	return v
}

// flatten implements the recursive expansion documented on [Add]. Nil
// values are dropped; everything else is stored under the dotted key.
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
