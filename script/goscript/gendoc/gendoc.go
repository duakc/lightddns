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

const I18nFileName = "i18n.yaml"

const WorkDirectory = "../../"

var RegisteredOption = map[string]struct{}{
	"Options":      {},
	"LogOption":    {},
	"DomainOption": {},

	// shared
	"ConnectOption": {},
	"DNSOption":     {},
	"HTTPOption":    {},

	// Datasource
	"DatasourceOption":              {},
	"CommandDatasourceOption":       {},
	"HTTPDatasourceOption":          {},
	"NetlinkDatasourceOption":       {},
	"DatasourceGroupFailoverOption": {},
	"DatasourceGroupSumOption":      {},

	// Provider
	"ProviderOption":           {},
	"CloudflareProviderOption": {},
}

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

// Run
// Deprecated: generate a doc is meaning less
func Run(ctx context.Context) {
	packageLoaded, err := gopackage.Load(goPackageConfigReplaceContext(ctx))
	if err != nil {
		Logger.Fatalf("Load Package: %s", err.Error())
	}
	structs := make([]*StructDocument, 0, len(RegisteredOption))
	for i := 0; i < len(packageLoaded); i++ {
		pkg := packageLoaded[i]
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				Logger.Errorf("package %s: %s", pkg.PkgPath, e)
			}
			continue
		}
		for j := 0; j < len(pkg.Syntax); j++ {
			structs = append(structs, handleFiles(pkg.Syntax[j])...)
		}
	}
}

func handleFiles(astFile *goast.File) []*StructDocument {
	d := make([]*StructDocument, 0)
	for i := 0; i < len(astFile.Decls); i++ {
		genDecl, ok := astFile.Decls[i].(*goast.GenDecl)
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
			if _, ok := RegisteredOption[typeSpec.Name.Name]; !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*goast.StructType)
			if !ok {
				continue
			}
			structDoc, err := NewStruct(genDecl, typeSpec, structType)
			if err != nil {
				Logger.Fatal(err)
			}
			d = append(d, structDoc)
		}
	}
	return d
}

func goPackageConfigReplaceContext(ctx context.Context) *gopackage.Config {
	if ctx == nil {
		panic("nil context")
	}
	gc := *GoPackageConfig
	gc.Context = ctx
	return &gc
}
