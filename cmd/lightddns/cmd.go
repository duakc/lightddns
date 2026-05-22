package main

import (
	"github.com/duakc/lightddns/cmd/lightddns/internal/check"
	"github.com/duakc/lightddns/cmd/lightddns/internal/globalcontext"
	"github.com/duakc/lightddns/cmd/lightddns/internal/run"
	"github.com/duakc/lightddns/cmd/lightddns/internal/version"
	"github.com/duakc/lightddns/constant"
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
	Use:              constant.Project,
	Short:            constant.Project + " is a lightweight dynamic DNS updater",
	PersistentPreRun: preRun,
}

var (
	workingDirectory string
	helper           filehelper.Helper
)

func init() {
	rootCommand.Flags().StringVarP(&workingDirectory, "workdir", "D", ".", "Working directory")

	rootCommand.AddCommand(
		run.Command,
		version.Command,
		check.Command)
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
}
