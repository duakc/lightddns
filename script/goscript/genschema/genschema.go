package genschema

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	_ "github.com/duakc/lightddns/datasources"
	"github.com/duakc/lightddns/infra/zaplog"
	_ "github.com/duakc/lightddns/providers"
	"github.com/duakc/lightddns/script/goscript/pkg/common"
	_ "github.com/duakc/lightddns/services"

	"go.uber.org/zap"
)

var output string

func init() {
	// Tracked in release/ so it resolves via raw.githubusercontent at a ref.
	flag.StringVar(&output, "o", common.ReleaseDir("schema.json"), "Output")
}

func Run(ctx context.Context) {
	flag.Parse()
	schema, err := GenSchema()
	if err != nil {
		zaplog.Fatal("GenSchema", zap.Error(err))
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		zaplog.Fatal("Mkdir", zap.Error(err))
	}
	if err := os.WriteFile(output, schema, 0o666); err != nil {
		zaplog.Fatal("WriteFile", zap.Error(err))
	}
}
