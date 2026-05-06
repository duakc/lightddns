package gendoc

import (
	"net/http"

	"github.com/duakc/lightddns/script/goscript/gendoc/jsonschema"
)

const (
	JSONTypeNull    = "null"
	JSONTypeNumber  = "number"
	JSONTypeBoolean = "boolean"
	JSONTypeInteger = "integer"
	JSONTypeString  = "string"
	JSONTypeArray   = "array"
	JSONTypeObject  = "object"
)

func stringOr(r *jsonschema.Schema) *jsonschema.Schema {
	return &jsonschema.Schema{AnyOf: []*jsonschema.Schema{
		singleType(JSONTypeString),
		r,
	}}
}

func singleType(r string) *jsonschema.Schema {
	return &jsonschema.Schema{Type: r}
}

func listAble(r string) *jsonschema.Schema {
	return &jsonschema.Schema{AnyOf: []*jsonschema.Schema{
		singleType(r),
		{
			Type:  JSONTypeArray,
			Items: singleType(r),
		},
	}}
}

func httpMethod() *jsonschema.Schema {
	return enumSchema(JSONTypeString,
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodConnect,
		http.MethodHead, http.MethodOptions, http.MethodTrace, http.MethodPatch,
		http.MethodDelete,
		"BREW", "PROPFIND", "WHEN")
}

func enumSchema(r string, vv ...any) *jsonschema.Schema {
	return &jsonschema.Schema{Type: r, Enum: vv}
}
