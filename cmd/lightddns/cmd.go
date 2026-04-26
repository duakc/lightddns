package main

import (
	"github.com/duakc/lightddns/cmd/lightddns/internal/run"
	"github.com/duakc/lightddns/cmd/lightddns/internal/version"
	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCommand = &cobra.Command{
	Use: "lightddns",
}

func init() {
	rootCommand.AddCommand(
		run.Command,
		version.Command)
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		zaplog.Fatal("Execute failed", zap.Error(err))
	}
}
