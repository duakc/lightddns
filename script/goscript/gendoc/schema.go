package gendoc

import (
	"bufio"
	"context"
	goast "go/ast"
	gotoken "go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/duakc/lightddns/infra/zaplog"
	gopackage "golang.org/x/tools/go/packages"
)

type BaseDoc struct {
	Name string
	Lang map[LangCode]string
}

type OptionFieldDoc struct {
	BaseDoc

	Required bool
	Values   []string

	YAML string
}

func (o *OptionFieldDoc) FromComment(comment *goast.CommentGroup) {
	scanner := bufio.NewScanner(strings.NewReader(comment.Text()))
	for scanner.Scan() {

	}
}

type OptionDoc struct {
	BaseDoc

	Fields []OptionFieldDoc
}

func NewOptionDoc(typeSpec *goast.TypeSpec, structType *goast.StructType) (*OptionDoc, error) {

	doc := &OptionDoc{}
	doc.Name = typeSpec.Name.Name

	for i := 0; i < len(structType.Fields.List); i++ {
		field := structType.Fields.List[i]
		if field.Doc == nil {
			// skip
			continue
		}
		fieldDoc := &OptionFieldDoc{}
		if field.Tag != nil {
			structTag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			fieldDoc.YAML = structTag.Get("yaml")
		}
		fieldDoc.FromComment(field.Doc)
	}
	return nil, nil
}

const WorkDirectory = "../../"

var (
	Logger = zaplog.NewPackage("goscript", "gendoc").Sugar()

	OptionDirectory = filepath.Join(WorkDirectory, "./options")

	GoPackageConfig = &gopackage.Config{
		Mode:       gopackage.LoadSyntax,
		Dir:        OptionDirectory,
		BuildFlags: []string{"debug"},
		Context:    context.Background(),
		Logf:       Logger.Infof,
		Env:        os.Environ(),
		Fset:       gotoken.NewFileSet(),
	}
)

func Run(ctx context.Context) {
	packageLoaded, err := gopackage.Load(goPackageConfigReplaceContext(ctx))
	if err != nil {
		Logger.Error("Load Package: %s", err.Error())
	}
	for i := 0; i < len(packageLoaded); i++ {
		pkg := packageLoaded[i]
		for i := 0; i < len(pkg.Syntax); i++ {
			handleFiles(pkg.Syntax[i])
		}
	}
}

func handleFiles(astFile *goast.File) {
	for i := 0; i < len(astFile.Decls); i++ {
		decl := astFile.Decls[i]
		genDecl, ok := decl.(*goast.GenDecl)
		// we only handle type here
		if !ok || genDecl.Tok != gotoken.TYPE {
			continue
		}
		for j := 0; j < len(genDecl.Specs); j++ {
			spec := genDecl.Specs[j]
			typeSpec, ok := spec.(*goast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*goast.StructType)
			if !ok {
				continue
			}
			NewOptionDoc(typeSpec, structType)
		}
	}
}

func goPackageConfigReplaceContext(ctx context.Context) *gopackage.Config {
	if ctx == nil {
		panic("nil context")
	}
	gc := *GoPackageConfig
	gc.Context = ctx
	return &gc
}
