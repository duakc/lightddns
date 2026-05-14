package genschema

import (
	"context"
	"flag"
	"os"

	_ "github.com/duakc/lightddns"
	"github.com/duakc/lightddns/infra/zaplog"

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
	if err := os.WriteFile(output, schema, 0o666); err != nil {
		zaplog.Fatal("WriteFile", zap.Error(err))
	}
}
