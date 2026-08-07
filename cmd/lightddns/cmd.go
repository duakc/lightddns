package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/duakc/lightddns/cmd/lightddns/internal/check"
	"github.com/duakc/lightddns/cmd/lightddns/internal/common"
	"github.com/duakc/lightddns/cmd/lightddns/internal/run"
	"github.com/duakc/lightddns/cmd/lightddns/internal/version"
	constpkg "github.com/duakc/lightddns/constant"
	// registry
	_ "github.com/duakc/lightddns/datasources"
	"github.com/duakc/lightddns/infra/zaplog"
	_ "github.com/duakc/lightddns/providers"
	_ "github.com/duakc/lightddns/services"

	"github.com/duakc/mt/services"
	"github.com/duakc/mt/services/closeme"
	"github.com/duakc/mt/services/container"
	"github.com/duakc/mt/services/filehelper"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var closeManager closeme.Manager

var rootCommand = &cobra.Command{
	Use:              constpkg.Project,
	Short:            constpkg.ProjectDescription,
	PersistentPreRun: preRun,
}

var (
	workingDirectory string
	envFile          string
)

func init() {
	rootCommand.PersistentFlags().StringVarP(&workingDirectory, "workdir", "D", ".", "Working directory")
	rootCommand.PersistentFlags().StringVar(&envFile, "env-file", "",
		"Load KEY=VALUE pairs from this file into the environment ({{ .Env }}); "+
			"default: use the inherited environment only")

	rootCommand.AddCommand(
		run.Command,
		version.Command,
		check.Command)
}

func preRun(cmd *cobra.Command, args []string) {
	ctx := services.NewRegistry(common.Context(), services.NewDefaultRegistry())

	closeManager = closeme.NewManager()

	fileHelper, err := filehelper.New(workingDirectory)
	if err != nil {
		zaplog.Fatal("create working directory failed", zap.Error(err))
	}
	defer closeme.AddClose(closeManager, fileHelper)

	// Opt-in: fold an env file into the process environment so every command
	// (and {{ .Env }} in the config) sees it.
	if envFile != "" {
		if err := common.ApplyEnvFile(fileHelper, envFile); err != nil {
			zaplog.Fatal("apply env file failed", zap.String("file", envFile), zap.Error(err))
		}
	}

	services.Store[filehelper.Helper](ctx, fileHelper)
	services.Store[closeme.Manager](ctx, closeManager)
	services.Store[container.Provider](ctx, container.NewDefaultProvider())

	common.StoreContext(ctx)
}

func main() {
	defer func() {
		var closeErr error
		if closeManager != nil {
			closeErr = closeManager.Close()
		}
		if err := errors.Join(closeErr, closeme.Default.Close()); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "close resources failed:\n%v", err)
		}
	}()
	if err := rootCommand.Execute(); err != nil {
		zaplog.Fatal("execute failed", zap.Error(err))
	}
}
