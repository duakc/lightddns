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

func Run(ctx context.Context) {
	packageLoaded, err := gopackage.Load(goPackageConfigReplaceContext(ctx))
	if err != nil {
		Logger.Fatalf("Load Package: %s", err.Error())
	}
	structs := make([]*StructDocument, 0, len(RegisteredOption))
	for i := 0; i < len(packageLoaded); i++ {
		pkg := packageLoaded[i]
		for i := 0; i < len(pkg.Syntax); i++ {
			structs = append(structs, handleFiles(pkg.Syntax[i])...)
		}
	}
	var schema []byte
	schema, err = GenSchema(structs)
	if err != nil {
		Logger.Fatal(err)
	}
	os.WriteFile("JSONSchema.json", schema, 0o666)
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
