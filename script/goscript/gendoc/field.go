package gendoc

import (
	"fmt"
	goast "go/ast"
	"reflect"
	"strings"

	"github.com/duakc/lightddns/script/goscript/pkg/jsonschema"
)

type FieldDocument struct {
	BaseDocument

	Ast *goast.Field

	YAML     *jsonschema.FieldTagInfo
	Values   []string
	Platform []string
	Shared   bool
}

func NewField(ast *goast.Field) (*FieldDocument, error) {
	o := &FieldDocument{}
	o.Ast = ast
	o.Name = identNames(ast.Names)
	if ast.Tag != nil {
		structTag := reflect.StructTag(strings.Trim(ast.Tag.Value, "`"))
		yamlInfo := structTag.Get("yaml")
		info := &jsonschema.FieldTagInfo{Name: o.Name}
		info.NewTag(yamlInfo)
		o.YAML = info
	}
	err := o.FromComment(ast.Doc, o)
	if err != nil {
		return nil, fmt.Errorf("field: %s: %w", o.Name, err)
	}
	return o, nil
}

func (o *FieldDocument) Required() bool {
	if o.YAML == nil || o.YAML.Omit || o.Shared {
		return false
	}
	return !o.YAML.Settings["omitempty"] && !o.YAML.Settings["omitzero"]
}

func (o *FieldDocument) FromTokenName(name string, t *Tokenizer) error {
	switch name {
	case "":
	case "@Shared":
		o.Shared = true
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
