package gendoc

import (
	"fmt"
	goast "go/ast"
	"reflect"
	"slices"
	"strings"
)

type FieldDocument struct {
	BaseDocument

	Ast *goast.Field

	YAML string

	Required bool
	Values   []string
}

func NewField(ast *goast.Field) (*FieldDocument, error) {
	o := &FieldDocument{}
	o.Ast = ast
	o.Name = identNames(ast.Names)
	o.Lang = make(map[LangCode]string)
	if ast.Tag != nil {
		structTag := reflect.StructTag(strings.Trim(ast.Tag.Value, "`"))
		o.YAML = structTag.Get("yaml")
	}
	err := o.FromComment(ast.Doc, o.FromTokenName)
	if err != nil {
		return nil, fmt.Errorf("field: %s: %w", o.Name, err)
	}
	return o, nil
}

func (o *FieldDocument) FromTokenName(name string, t *Tokenizer) error {
	switch name {
	case "":
	case "@Required":
		o.Required = true
	case "@Values":
		t.NextMeta()
		values := strings.Split(t.FragmentText(), ",")
		if len(values) == 0 {
			return fmt.Errorf("%s: %w", name, ErrMissingMetaValue)
		}
		o.Values = slices.DeleteFunc(values, func(s string) bool {
			return len(s) == 0
		})
	default:
		return fmt.Errorf("unknown meta: %s", name)
	}
	return nil
}
