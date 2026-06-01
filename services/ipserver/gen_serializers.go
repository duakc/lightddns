//go:build generate

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
)

type Field struct {
	Name    string
	Type    string
	JSONKey string
	YAMLKey string
}

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "response.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			var fields []Field
			for _, field := range structType.Fields.List {
				tag := parseTag(field.Tag)
				for _, name := range field.Names {
					jsonKey := tagKey(tag.Get("json"), name.Name)
					yamlKey := tagKey(tag.Get("yaml"), name.Name)
					fields = append(fields, Field{
						Name:    name.Name,
						Type:    typeToStr(field.Type),
						JSONKey: jsonKey,
						YAMLKey: yamlKey,
					})
				}
			}
			writeJSONFile(typeSpec.Name.Name, fields)
			writeYAMLFile(typeSpec.Name.Name, fields)
		}
	}
}

func parseTag(lit *ast.BasicLit) reflect.StructTag {
	if lit == nil {
		return ""
	}
	return reflect.StructTag(strings.Trim(lit.Value, "`"))
}

func tagKey(raw, fieldName string) string {
	if raw == "" {
		return strings.ToLower(fieldName)
	}
	name, _, _ := strings.Cut(raw, ",")
	if name == "-" {
		return ""
	}
	if name == "" {
		return strings.ToLower(fieldName)
	}
	return name
}

const (
	avgStringValue = 24 // "<~22 chars>"
	avgBoolValue   = 5  // "false"
	avgOtherValue  = 18 // "<~16 chars>" via fmt.Sprint
)

func estimateJSONSize(fields []Field) int {
	size := 2 // braces
	n := 0
	for _, f := range fields {
		if f.JSONKey == "" {
			continue
		}
		n++
		size += len(f.JSONKey) + 3 // "key":
		size += valueCost(f.Type)
	}
	if n > 1 {
		size += n - 1 // commas
	}
	return size
}

func estimateYAMLSize(fields []Field) int {
	size := 0
	for _, f := range fields {
		if f.YAMLKey == "" {
			continue
		}
		size += len(f.YAMLKey) + 2 // "key: "
		size += valueCost(f.Type)
		size++ // newline
	}
	return size
}

func valueCost(typ string) int {
	switch typ {
	case "string":
		return avgStringValue
	case "bool":
		return avgBoolValue
	default:
		return avgOtherValue
	}
}

type codeBuf struct {
	bytes.Buffer
}

func (b *codeBuf) line(format string, args ...any) {
	b.WriteByte('\t')
	if len(args) == 0 {
		b.WriteString(format)
	} else {
		_, _ = fmt.Fprintf(&b.Buffer, format, args...)
	}
	b.WriteByte('\n')
}

func (b *codeBuf) raw(s string) { b.WriteString(s) }

func emitStringField(b *codeBuf, fieldName string) {
	b.line(`buffer.WriteByte('"')`)
	b.line("buffer.WriteString(resp.%s)", fieldName)
	b.line(`buffer.WriteByte('"')`)
}

func emitBoolField(b *codeBuf, fieldName string) {
	b.line("if resp.%s {", fieldName)
	b.line("\tbuffer.WriteString(\"true\")")
	b.line("} else {")
	b.line("\tbuffer.WriteString(\"false\")")
	b.line("}")
}

func emitFmtField(b *codeBuf, fieldName string, quoted bool) {
	if quoted {
		b.line(`buffer.WriteByte('"')`)
	}
	b.line("_, _ = fmt.Fprint(buffer, resp.%s)", fieldName)
	if quoted {
		b.line(`buffer.WriteByte('"')`)
	}
}

func writeJSONFile(structName string, fields []Field) {
	var body codeBuf
	needsFmt := false

	_, _ = fmt.Fprintf(&body, "func (resp %s) writeJSON(buffer *bytes.Buffer) {\n", structName)
	body.line("buffer.Grow(%d) // estimated size: keys + avg value cost", estimateJSONSize(fields))
	body.line(`buffer.WriteByte('{')`)

	first := true
	for _, field := range fields {
		if field.JSONKey == "" {
			continue
		}
		if !first {
			body.line(`buffer.WriteByte(',')`)
		}
		first = false

		body.line("buffer.WriteString(`\"%s\":`)", field.JSONKey)

		switch field.Type {
		case "string":
			emitStringField(&body, field.Name)
		case "bool":
			emitBoolField(&body, field.Name)
		default:
			needsFmt = true
			emitFmtField(&body, field.Name, true)
		}
	}

	body.line(`buffer.WriteByte('}')`)
	body.raw("}\n")

	writeGenFile(strings.ToLower(structName)+"_json.go", needsFmt, body.Bytes())
}

func writeYAMLFile(structName string, fields []Field) {
	var body codeBuf
	needsFmt := false

	_, _ = fmt.Fprintf(&body, "func (resp %s) writeYAML(buffer *bytes.Buffer) {\n", structName)
	body.line("buffer.Grow(%d) // estimated size: keys + avg value cost", estimateYAMLSize(fields))

	for _, field := range fields {
		if field.YAMLKey == "" {
			continue
		}
		body.line("buffer.WriteString(`%s: `)", field.YAMLKey)

		switch field.Type {
		case "string":
			emitStringField(&body, field.Name)
		case "bool":
			emitBoolField(&body, field.Name)
		default:
			needsFmt = true
			emitFmtField(&body, field.Name, false)
		}

		body.line(`buffer.WriteByte('\n')`)
	}

	body.raw("}\n")

	writeGenFile(strings.ToLower(structName)+"_yaml.go", needsFmt, body.Bytes())
}

func writeGenFile(fileName string, needsFmt bool, body []byte) {
	var code bytes.Buffer
	code.WriteString("// Code generated by gen_serializers.go; DO NOT EDIT.\n\n")
	code.WriteString("package ipserver\n\n")
	code.WriteString("import (\n\t\"bytes\"\n")
	if needsFmt {
		code.WriteString("\t\"fmt\"\n")
	}
	code.WriteString(")\n\n")
	code.Write(body)

	if err := os.WriteFile(fileName, code.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
}

func typeToStr(expr ast.Expr) string {
	ident, ok := expr.(*ast.Ident)
	if ok {
		return ident.Name
	}
	return ""
}
