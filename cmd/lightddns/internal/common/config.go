package common

import (
	"bytes"
	"context"
	"html/template"
	"path/filepath"

	"github.com/duakc/lightddns/options"

	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/filehelper"

	goyaml "github.com/goccy/go-yaml"
)

// LoadConfig reads a config file, expands it as a Go template exposing the
// process environment as {{ .Env.KEY }}, and strictly decodes it into Options.
func LoadConfig(ctx context.Context, file string) (options.Options, error) {
	var opt options.Options

	// An absolute path is read as-is; a relative one is resolved against the
	// working directory (-D). This keeps config and state dirs decoupled.
	path := file
	if !filepath.IsAbs(path) {
		fileHelper := services.Lookup[filehelper.Helper](ctx)
		path = fileHelper.Path(file)
	}

	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return opt, err
	}

	buffer := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buffer, struct{ Env map[string]string }{Env: envMap()}); err != nil {
		return opt, err
	}

	if err := goyaml.NewDecoder(buffer, goyaml.DisallowUnknownField()).Decode(&opt); err != nil {
		return opt, err
	}
	return opt, nil
}
