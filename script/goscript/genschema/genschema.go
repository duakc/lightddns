package genschema

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	_ "github.com/duakc/lightddns/datasources"
	"github.com/duakc/lightddns/infra/zaplog"
	_ "github.com/duakc/lightddns/providers"

	"go.uber.org/zap"
)

var output string

func init() {
	flag.StringVar(&output, "o", "./build/schema.json", "Output")
}

func Run(ctx context.Context) {
	flag.Parse()
	schema, err := GenSchema()
	if err != nil {
		zaplog.Fatal("GenSchema", zap.Error(err))
	}
	if err := os.Mkdir(filepath.Dir(output), 0o755); err != nil {
		zaplog.Fatal("Mkdir", zap.Error(err))
	}
	if err := os.WriteFile(output, schema, 0o666); err != nil {
		zaplog.Fatal("WriteFile", zap.Error(err))
	}
}
