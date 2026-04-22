package main

import (
	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCommand = &cobra.Command{
	Use: "lightddns",
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		zaplog.Fatal("Execute failed", zap.Error(err))
	}
}
