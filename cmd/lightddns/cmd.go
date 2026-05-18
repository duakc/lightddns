package main

import (
	"github.com/duakc/lightddns/cmd/lightddns/internal/run"
	"github.com/duakc/lightddns/cmd/lightddns/internal/version"
	// registry
	_ "github.com/duakc/lightddns/datasources"
	"github.com/duakc/lightddns/infra/filehelper"
	"github.com/duakc/lightddns/infra/zaplog"
	_ "github.com/duakc/lightddns/providers"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCommand = &cobra.Command{
	Use:              "lightddns",
	PersistentPreRun: preRun,
}

var workingDirectory string

func init() {
	rootCommand.Flags().StringVarP(&workingDirectory, "workdir", "D", ".", "working directory")

	rootCommand.AddCommand(
		run.Command,
		version.Command)
}

func preRun(cmd *cobra.Command, args []string) {
	filehelper.WorkingDir(workingDirectory)
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		zaplog.Fatal("Execute failed", zap.Error(err))
	}
}
