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
	"strconv"
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

const boolMaxLen = 5

type sizeExpr struct {
	constant int
	vars     []string
}

func (s *sizeExpr) addConst(n int)     { s.constant += n }
func (s *sizeExpr) addVar(expr string) { s.vars = append(s.vars, expr) }

func (s sizeExpr) String() string {
	parts := []string{strconv.Itoa(s.constant)}
	parts = append(parts, s.vars...)
	return strings.Join(parts, " + ")
}

func (s *sizeExpr) addValue(field Field) {
	switch field.Type {
	case "string":
		s.addConst(2) // surrounding quotes
		s.addVar(fmt.Sprintf("len(resp.%s)", field.Name))
	case "bool":
		s.addConst(boolMaxLen)
	}
}

func computeJSONSize(fields []Field) sizeExpr {
	var s sizeExpr
	s.addConst(2) // braces
	n := 0
	for _, f := range fields {
		if f.JSONKey == "" {
			continue
		}
		n++
		s.addConst(len(f.JSONKey) + 3) // "key":
		s.addValue(f)
	}
	if n > 1 {
		s.addConst(n - 1) // commas
	}
	return s
}

func computeYAMLSize(fields []Field) sizeExpr {
	var s sizeExpr
	for _, f := range fields {
		if f.YAMLKey == "" {
			continue
		}
		s.addConst(len(f.YAMLKey) + 2) // "key: "
		s.addValue(f)
		s.addConst(1) // newline
	}
	return s
}

type codeBuf struct {
	bytes.Buffer
}

func (b *codeBuf) linef(format string, args ...any) {
	b.WriteByte('\t')
	if len(args) == 0 {
		b.WriteString(format)
	} else {
		_, _ = fmt.Fprintf(&b.Buffer, format, args...)
	}
	b.WriteByte('\n')
}

func (b *codeBuf) raw(s string) { b.WriteString(s) }

func (b *codeBuf) emitByte(charLit string) {
	b.linef("buf.WriteByte(%s)", charLit)
	b.linef("n++")
}

func (b *codeBuf) emitLit(rawText string) {
	b.linef("buf.WriteString(`%s`)", rawText)
	b.linef("n += %d", len(rawText))
}

func emitStringField(b *codeBuf, fieldName string) {
	b.emitByte(`'"'`)
	b.linef("buf.WriteString(resp.%s)", fieldName)
	b.linef("n += len(resp.%s)", fieldName)
	b.emitByte(`'"'`)
}

func emitBoolField(b *codeBuf, fieldName string) {
	b.linef("if resp.%s {", fieldName)
	b.linef(`	buf.WriteString("true")`)
	b.linef("\tn += 4")
	b.linef("} else {")
	b.linef(`	buf.WriteString("false")`)
	b.linef("\tn += 5")
	b.linef("}")
}

func rejectField(structName string, f Field) {
	log.Fatalf("gen_serializers: %s.%s has unsupported type %q "+
		"(only string/bool are handled); extend the generator before adding this field",
		structName, f.Name, f.Type)
}

func writeJSONFile(structName string, fields []Field) {
	var body codeBuf

	_, _ = fmt.Fprintf(&body, "func (resp *%s) writeJSON(buf freebuf.Buffer) int {\n", structName)
	body.linef("buf.Grow(%s) // exact: constants + len(strings) + worst-case bool", computeJSONSize(fields))
	body.linef("n := 0")
	body.emitByte(`'{'`)

	first := true
	for _, field := range fields {
		if field.JSONKey == "" {
			continue
		}
		if !first {
			body.emitByte(`','`)
		}
		first = false

		body.emitLit(fmt.Sprintf(`"%s":`, field.JSONKey))

		switch field.Type {
		case "string":
			emitStringField(&body, field.Name)
		case "bool":
			emitBoolField(&body, field.Name)
		default:
			rejectField(structName, field)
		}
	}

	body.emitByte(`'}'`)
	body.linef("return n")
	body.raw("}\n")

	writeGenFile(strings.ToLower(structName)+"_json.go", body.Bytes())
}

func writeYAMLFile(structName string, fields []Field) {
	var body codeBuf

	_, _ = fmt.Fprintf(&body, "func (resp *%s) writeYAML(buf freebuf.Buffer) int {\n", structName)
	body.linef("buf.Grow(%s) // exact: constants + len(strings) + worst-case bool", computeYAMLSize(fields))
	body.linef("n := 0")

	for _, field := range fields {
		if field.YAMLKey == "" {
			continue
		}
		body.emitLit(fmt.Sprintf("%s: ", field.YAMLKey))

		switch field.Type {
		case "string":
			emitStringField(&body, field.Name)
		case "bool":
			emitBoolField(&body, field.Name)
		default:
			rejectField(structName, field)
		}

		body.emitByte(`'\n'`)
	}

	body.linef("return n")
	body.raw("}\n")

	writeGenFile(strings.ToLower(structName)+"_yaml.go", body.Bytes())
}

func writeGenFile(fileName string, body []byte) {
	var code bytes.Buffer
	code.WriteString("// Code generated by gen_serializers.go; DO NOT EDIT.\n\n")
	code.WriteString("package ipserver\n\n")
	code.WriteString("import (\n\t\"github.com/duakc/mt/freebuf\"\n)\n\n")
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
