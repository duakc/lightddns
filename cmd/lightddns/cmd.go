package main

import (
	"errors"

	"github.com/duakc/lightddns/cmd/lightddns/internal/globalcontext"
	"github.com/duakc/lightddns/cmd/lightddns/internal/run"
	"github.com/duakc/lightddns/cmd/lightddns/internal/version"
	// registry
	_ "github.com/duakc/lightddns/datasources"
	"github.com/duakc/lightddns/infra/zaplog"
	_ "github.com/duakc/lightddns/providers"

	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/filehelper"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCommand = &cobra.Command{
	Use:              "lightddns",
	PersistentPreRun: preRun,
}

var (
	workingDirectory string
	helper           filehelper.Helper
)

func init() {
	rootCommand.Flags().StringVarP(&workingDirectory, "workdir", "D", ".", "working directory")

	rootCommand.AddCommand(
		run.Command,
		version.Command)
}

func preRun(cmd *cobra.Command, args []string) {
	ctx := services.NewRegistry(globalcontext.Load(), services.NewDefaultRegistry())

	files, err := filehelper.NewMkdir(workingDirectory)
	if err != nil {
		zaplog.Fatal("create working directory failed", zap.Error(err))
	}
	helper = files
	services.Store(ctx, files)

	globalcontext.Store(ctx)
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		zaplog.Fatal("execute failed", zap.Error(err))
	}

	// resources clean
	if err := errors.Join(helper.Close()); err != nil {
		zaplog.Fatal("close failed", zap.Error(err))
	}
}
