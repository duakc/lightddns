package check

import (
	"github.com/duakc/lightddns/cmd/lightddns/internal/common"
	"github.com/duakc/lightddns/infra/zaplog"

	"github.com/duakc/lightddns"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type Arguments struct {
	Config string
}

var (
	commandArguments = &Arguments{}
	Command          = &cobra.Command{
		Use:   "check",
		Short: "Check configuration validity",
		Long:  "Parse the configuration file, create datasources and providers, then exit. Use this to validate config before deploying.",
		Run:   entry,
	}
)

func init() {
	Command.Flags().StringVarP(&commandArguments.Config, "config", "c", "", "Path to configuration file")
}

func entry(cmd *cobra.Command, args []string) {
	ctx := common.Context()

	// Same load path as `run`: template expansion + strict decoding, so a
	// passing check means the daemon will start with this exact config.
	option, err := common.LoadConfig(ctx, commandArguments.Config)
	if err != nil {
		zaplog.Fatal("load config failed", zap.Error(err))
	}

	var ddns *lightddns.LightDDNS
	if ddns, err = lightddns.New(ctx, option); err != nil {
		zaplog.Fatal("create new ddns instance failed", zap.Error(err))
	}

	_ = ddns.Close()

	zaplog.Info("configuration is valid", zap.String("file", commandArguments.Config))
}
