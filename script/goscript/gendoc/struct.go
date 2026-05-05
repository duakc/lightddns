package gendoc

import (
	"fmt"
	goast "go/ast"
)

type StructDocument struct {
	BaseDocument

	Ast         *goast.StructType
	AstTypeSpec *goast.TypeSpec

	Fields []*FieldDocument

	Shared bool
}

func (o *StructDocument) FromTokenName(name string, token *Tokenizer) error {
	switch name {
	case "@Shared":
		o.Shared = true
	default:
		return fmt.Errorf("unknown meta: %s", name)
	}
	return nil
}

func NewStruct(typeSpec *goast.TypeSpec, structType *goast.StructType) (*StructDocument, error) {
	o := &StructDocument{}
	o.AstTypeSpec = typeSpec
	o.Ast = structType
	o.Name = typeSpec.Name.Name
	err := o.FromComment(typeSpec.Doc, o.FromTokenName)
	if err != nil {
		return nil, err
	}
	// fields
	for i := 0; i < len(structType.Fields.List); i++ {
		field := structType.Fields.List[i]
		fd, err := NewField(field)
		if err != nil {
			return nil, err
		}
		o.Fields = append(o.Fields, fd)
	}

	return o, nil
}
