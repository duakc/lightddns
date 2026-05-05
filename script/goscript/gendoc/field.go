package gendoc

import (
	"fmt"
	goast "go/ast"
	"reflect"
	"strings"
)

type FieldDocument struct {
	BaseDocument

	Ast *goast.Field

	YAML string

	Required bool
	Values   []string
	Platform []string
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
	err := o.FromComment(ast.Doc, o)
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
		o.Values = docArray(t.FragmentText())
		if len(o.Values) == 0 {
			return fmt.Errorf("%s: %w", name, ErrMissingMetaValue)
		}
	case "@Platform":
		t.NextMeta()
		o.Platform = docArray(t.FragmentText())
		if len(o.Platform) == 0 {
			return fmt.Errorf("%s: %w", name, ErrMissingMetaValue)
		}
	default:
		return fmt.Errorf("unknown meta: %s", name)
	}
	return nil
}
