package gendoc

import (
	"context"
	goast "go/ast"
	gotoken "go/token"
	"os"
	"path/filepath"

	"github.com/duakc/lightddns/infra/zaplog"
	gopackage "golang.org/x/tools/go/packages"
)

const WorkDirectory = "../../"

var (
	Logger = zaplog.NewPackage("goscript", "gendoc").Sugar()

	OptionDirectory = filepath.Join(WorkDirectory, "./options")

	GoPackageConfig = &gopackage.Config{
		Mode:       gopackage.NeedName | gopackage.NeedFiles | gopackage.NeedSyntax,
		Dir:        OptionDirectory,
		BuildFlags: []string{"-tags=debug"},
		Context:    context.Background(),
		Logf:       Logger.Infof,
		Env:        os.Environ(),
		Fset:       gotoken.NewFileSet(),
	}
)

func Run(ctx context.Context) {
	packageLoaded, err := gopackage.Load(goPackageConfigReplaceContext(ctx))
	if err != nil {
		Logger.Fatalf("Load Package: %s", err.Error())
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
			NewStruct(typeSpec, structType)
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
