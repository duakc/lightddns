package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/duakc/lightddns/cmd/lightddns/internal/check"
	"github.com/duakc/lightddns/cmd/lightddns/internal/globalcontext"
	"github.com/duakc/lightddns/cmd/lightddns/internal/run"
	"github.com/duakc/lightddns/cmd/lightddns/internal/version"
	"github.com/duakc/lightddns/constant"
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
	Use:              constant.Project,
	Short:            constant.Project + " is a lightweight dynamic DNS updater",
	PersistentPreRun: preRun,
}

var workingDirectory string

func init() {
	rootCommand.Flags().StringVarP(&workingDirectory, "workdir", "D", ".", "Working directory")

	rootCommand.AddCommand(
		run.Command,
		version.Command,
		check.Command)
}

func preRun(cmd *cobra.Command, args []string) {
	ctx := services.NewRegistry(globalcontext.Load(), services.NewDefaultRegistry())

	closeManager = closeme.NewManager()

	fileHelper, err := filehelper.New(workingDirectory)
	if err != nil {
		zaplog.Fatal("create working directory failed", zap.Error(err))
	}
	defer closeme.AddClose(closeManager, fileHelper)

	services.Store[filehelper.Helper](ctx, fileHelper)
	services.Store[closeme.Manager](ctx, closeManager)
	services.Store[container.Provider](ctx, container.NewDefaultProvider())

	globalcontext.Store(ctx)
}

func main() {
	defer func() {
		if err := errors.Join(closeManager.Close(), closeme.Default.Close()); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "close resources failed:\n%v", err)
		}
	}()
	if err := rootCommand.Execute(); err != nil {
		zaplog.Fatal("execute failed", zap.Error(err))
	}
}
